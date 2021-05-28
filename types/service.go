package types

import (
	"context"

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
// The arg should be of type diff.Event, containing the service to be created,
// else the function will panic.
// It returns a the created *state.Service.
func (s *serviceCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	service := serviceFromStruct(event)
	createdService, err := s.client.Services.Create(ctx, &service.Service)
	if err != nil {
		return nil, err
	}
	return &state.Service{Service: *createdService}, nil
}

// Delete deletes a Service in Kong.
// The arg should be of type diff.Event, containing the service to be deleted,
// else the function will panic.
// It returns a the deleted *state.Service.
func (s *serviceCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	service := serviceFromStruct(event)
	err := s.client.Services.Delete(ctx, service.ID)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// Update updates a Service in Kong.
// The arg should be of type diff.Event, containing the service to be updated,
// else the function will panic.
// It returns a the updated *state.Service.
func (s *serviceCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	service := serviceFromStruct(event)

	updatedService, err := s.client.Services.Create(ctx, &service.Service)
	if err != nil {
		return nil, err
	}
	return &state.Service{Service: *updatedService}, nil
}

type servicePostAction struct {
	// TODO switch back to unexported
	CurrentState *state.KongState
}

func (crud *servicePostAction) Create(ctx context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.CurrentState.Services.Add(*args[0].(*state.Service))
}

func (crud *servicePostAction) Delete(ctx context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.CurrentState.Services.Delete(*((args[0].(*state.Service)).ID))
}

func (crud *servicePostAction) Update(ctx context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.CurrentState.Services.Update(*args[0].(*state.Service))
}
