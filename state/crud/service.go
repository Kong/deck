package crud

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
)

type ServiceCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func argsForService(arg ...crud.Arg) (*state.Service, *state.KongState, *state.KongState, *kong.Client) {
	service, ok := arg[0].(*state.Service)
	if !ok {
		panic("unexpected type, expected *state.Service")
	}
	currentState, ok := arg[1].(*state.KongState)
	if !ok {
		panic("unexpected type, expected *state.KongState")
	}
	targetState, ok := arg[2].(*state.KongState)
	if !ok {
		panic("unexpected type, expected *state.KongState")
	}
	client, ok := arg[3].(*kong.Client)
	if !ok {
		panic("unexpected type, expected *kong.Client")
	}

	return service, currentState, targetState, client
}

func (s *ServiceCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	service, current, _, client := argsForService(arg...)
	createdService, err := client.Services.Create(nil, &service.Service)
	if err != nil {
		return nil, err
	}
	err = current.AddService(state.Service{Service: *createdService})
	if err != nil {
		return nil, err //TODO annotate error
	}
	return createdService, nil
}
func (s *ServiceCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	service, current, _, client := argsForService(arg...)
	err := client.Services.Delete(nil, service.ID)
	if err != nil {
		return nil, err
	}
	err = current.DeleteService(*service.ID)
	if err != nil {
		return nil, err //TODO annotate error
	}
	return nil, err
}
func (s *ServiceCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	service, current, _, client := argsForService(arg...)
	updatedService, err := client.Services.Update(nil, &service.Service)
	if err != nil {
		return nil, err
	}
	err = current.UpdateService(*service)
	if err != nil {
		return nil, err //TODO annotate error
	}
	return updatedService, nil
}

type RouteCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}
