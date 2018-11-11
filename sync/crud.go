package sync

import (
	"fmt"

	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
)

type Callback func(crud.Arg, error) error

type ServiceCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func args(arg ...crud.Arg) (*state.Service, *state.KongState, *state.KongState, *kong.Client) {
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
	service, current, _, client := args(arg...)
	createdService, err := client.Services.Create(nil, &service.Service)
	if err != nil {
		return nil, err
	}
	err = current.AddService(state.Service{*createdService})
	if err != nil {
		return nil, err //TODO annotate error
	}
	return createdService, nil
}
func (s *ServiceCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	service, current, _, client := args(arg...)
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
	service, current, _, client := args(arg...)
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

func args2(arg ...crud.Arg) (*state.Route, *state.KongState, *state.KongState, *kong.Client) {
	route, ok := arg[0].(*state.Route)
	if !ok {
		panic("unexpected type, expected *state.Route")
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

	return route, currentState, targetState, client
}

func (s *RouteCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	route, _, _, client := args2(arg...)
	fmt.Println("Create called with ", route)
	createdService, err := client.Routes.Create(nil, &route.Route)
	if err != nil {
		return nil, err
	}
	fmt.Println("created: ", createdService)
	return createdService, nil
}
func (s *RouteCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	route, _, _, client := args2(arg...)
	fmt.Println("Delete called with ", route)
	err := client.Routes.Delete(nil, route.ID)
	return nil, err
}
func (s *RouteCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	route, _, _, client := args2(arg...)
	fmt.Println("Update called with ", route)
	updatedService, err := client.Routes.Update(nil, &route.Route)
	if err != nil {
		return nil, err
	}
	return updatedService, nil
}
