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

	updatedService, err := s.client.Services.Update(ctx, &service.Service)
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
	if err == state.ErrNotFound {
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
		event, err := d.createUpdateService(service)
		if err != nil {
			return err
		}
		if event != nil {
			err = handler(*event)
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
	currentService, err := d.currentState.Services.Get(*targetService.Name)
	if err == state.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up service %q: %w",
			*targetService.Name, err)
	}

	if *currentService.ID != *targetService.ID {
		// Found a duplicate, delete it along with all routes and plugins associated with it.
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
			// We have to delete all plugins associated with the route to make sure they'll be recreated eventually.
			// Plugins are deleted by the cascading delete of the route and without us generating a delete event manually,
			// they could not be later recreated in createUpdates stage of the diff.
			// By generating a delete event for each plugin, we make sure that the implicit deletion of plugins is handled
			// in the local state and createUpdate stage can recreate them.
			pluginsToDelete, err := d.currentState.Plugins.GetAllByRouteID(*route.ID)
			if err != nil {
				return nil, fmt.Errorf("error looking up plugins for route %q: %w",
					*route.Name, err)
			}

			for _, plugin := range pluginsToDelete {
				if err := d.currentState.Plugins.Delete(*plugin.ID); err != nil {
					return nil, fmt.Errorf("error deleting plugin %q for route %q: %w",
						*route.Name, *plugin.Name, err)
				}
				// REVIEW: Should we generate a delete event for the plugin? It will always result in a DELETE
				// call for this plugin while it could be just a local state change as we already know it's going
				// to be deleted by the cascading delete of the route.
				//
				// It's also problematic when syncing with Konnect as Koko returns 404 when trying to delete a plugin
				// that doesn't exist. It's not an issue when syncing with Kong as Kong returns 204.
				//
				//events = append(events, crud.Event{
				//	Op:   crud.Delete,
				//	Kind: "plugin",
				//	Obj:  plugin,
				//})
			}

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
