package types

import (
	"context"
	"errors"
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
	kind crud.Kind

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
	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  service,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up service %q: %w",
			service.FriendlyName(), err)
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

	if errors.Is(err, state.ErrNotFound) {
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

func (d *serviceDiffer) DuplicatesDeletes() ([]crud.Event, error) {
	targetServices, err := d.targetState.Services.GetAll()
	if err != nil {
		return nil, fmt.Errorf("error fetching services from state: %w", err)
	}
	var events []crud.Event
	for _, service := range targetServices {
		serviceEvents, err := d.deleteDuplicateService(service)
		if err != nil {
			return nil, err
		}
		events = append(events, serviceEvents...)
	}

	return events, nil
}

func (d *serviceDiffer) deleteDuplicateService(targetService *state.Service) ([]crud.Event, error) {
	if targetService == nil || targetService.Name == nil {
		// Nothing to do, cannot be a duplicate with no name.
		return nil, nil
	}

	currentService, err := d.currentState.Services.Get(*targetService.Name)
	if errors.Is(err, state.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up service %q: %w",
			*targetService.Name, err)
	}

	if *currentService.ID != *targetService.ID {
		var events []crud.Event

		// We have to delete all routes beforehand as otherwise we will get a foreign key error when deleting the service
		// as routes are not deleted by the cascading delete of the service.
		// See https://github.com/Kong/kong/discussions/7314 for more details.
		routesToDelete, err := d.currentState.Routes.GetAllByServiceID(*currentService.ID)
		if err != nil {
			return nil, fmt.Errorf("error looking up routes for service %q: %w",
				*currentService.Name, err)
		}

		for _, route := range routesToDelete {
			events = append(events, crud.Event{
				Op:   crud.Delete,
				Kind: "route",
				Obj:  route,
			})
		}

		return append(events, crud.Event{
			Op:   crud.Delete,
			Kind: "service",
			Obj:  currentService,
		}), nil
	}

	return nil, nil
}
