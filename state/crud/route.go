package crud

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
)

func argsFroRoute(arg ...crud.Arg) (*state.Route, *state.KongState, *state.KongState, *kong.Client) {
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
	route, _, _, client := argsFroRoute(arg...)
	createdService, err := client.Routes.Create(nil, &route.Route)
	if err != nil {
		return nil, err
	}
	return createdService, nil
}
func (s *RouteCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	route, _, _, client := argsFroRoute(arg...)
	err := client.Routes.Delete(nil, route.ID)
	return nil, err
}
func (s *RouteCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	route, _, _, client := argsFroRoute(arg...)
	updatedService, err := client.Routes.Update(nil, &route.Route)
	if err != nil {
		return nil, err
	}
	return updatedService, nil
}
