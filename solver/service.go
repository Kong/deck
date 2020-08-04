package solver

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
	"github.com/hbagdi/go-kong/kong"
)

// serviceCRUD implements crud.Actions interface.
type serviceCRUD struct {
	client *kong.Client
}

func serviceFromStuct(arg diff.Event) *state.Service {
	service, ok := arg.Obj.(*state.Service)
	if !ok {
		panic("unexpected type, expected *state.service")
	}
	return service
}

// Create creates a Service in Kong.
// The arg should be of type diff.Event, containing the service to be created,
// else the function will panic.
// It returns a the created *state.Service.
func (s *serviceCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	service := serviceFromStuct(event)
	createdService, err := s.client.Services.Create(nil, &service.Service)
	if err != nil {
		return nil, err
	}
	return &state.Service{Service: *createdService}, nil
}

// Delete deletes a Service in Kong.
// The arg should be of type diff.Event, containing the service to be deleted,
// else the function will panic.
// It returns a the deleted *state.Service.
func (s *serviceCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	service := serviceFromStuct(event)
	err := s.client.Services.Delete(nil, service.ID)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// Update updates a Service in Kong.
// The arg should be of type diff.Event, containing the service to be updated,
// else the function will panic.
// It returns a the updated *state.Service.
func (s *serviceCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	service := serviceFromStuct(event)

	updatedService, err := s.client.Services.Create(nil, &service.Service)
	if err != nil {
		return nil, err
	}
	return &state.Service{Service: *updatedService}, nil
}
