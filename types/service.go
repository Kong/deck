package types

import (
	"context"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// serviceCRUD implements crud.Actions interface.
type serviceCRUD struct {
	client *kong.Client
}

func serviceFromStruct(arg crud.Event) *state.Service {
	service, ok := arg.Obj.(*state.Service)
	if !ok {
		panic("unexpected type, expected *state.service")
	}
	return service
}

// Create creates a Service in Kong.
// The arg should be of type crud.Event, containing the service to be created,
// else the function will panic.
// It returns a the created *state.Service.
func (s *serviceCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	service := serviceFromStruct(event)
	createdService, err := s.client.Services.Create(ctx, &service.Service)
	if err != nil {
		return nil, err
	}
	return &state.Service{Service: *createdService}, nil
}

// Delete deletes a Service in Kong.
// The arg should be of type crud.Event, containing the service to be deleted,
// else the function will panic.
// It returns a the deleted *state.Service.
func (s *serviceCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	service := serviceFromStruct(event)
	err := s.client.Services.Delete(ctx, service.ID)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// Update updates a Service in Kong.
// The arg should be of type crud.Event, containing the service to be updated,
// else the function will panic.
// It returns a the updated *state.Service.
func (s *serviceCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	service := serviceFromStruct(event)

	updatedService, err := s.client.Services.Create(ctx, &service.Service)
	if err != nil {
		return nil, err
	}
	return &state.Service{Service: *updatedService}, nil
}

type serviceDiffer struct {
	currentState, targetState *state.KongState
}

func (d *serviceDiffer) Deletes(handler func(crud.Event) error) error {
	currentServices, err := d.currentState.Services.GetAll()
	if err != nil {
		return err
	}

	for _, service := range currentServices {
		n, err := d.deleteService(service)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (d *serviceDiffer) deleteService(service *state.Service) (*crud.Event, error) {
	_, err := d.targetState.Services.Get(*service.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "service",
			Obj:  service,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up service %q: %w",
			service.Identifier(), err)
	}
	return nil, nil
}

func (d *serviceDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetServices, err := d.targetState.Services.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching services from state: %w", err)
	}

	for _, service := range targetServices {
		n, err := d.createUpdateService(service)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *serviceDiffer) createUpdateService(service *state.Service) (*crud.Event, error) {
	serviceCopy := &state.Service{Service: *service.DeepCopy()}
	currentService, err := d.currentState.Services.Get(*service.ID)

	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Create,
			Kind: "service",
			Obj:  serviceCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up service %q: %w",
			*service.Name, err)
	}

	// found, check if update needed
	if !currentService.EqualWithOpts(serviceCopy, false, true) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "service",
			Obj:    serviceCopy,
			OldObj: currentService,
		}, nil
	}
	return nil, nil
}
