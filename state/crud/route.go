package crud

import (
	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

// RouteCRUD implements Actions interface
// from the github.com/kong/crud package for the Route entitiy of Kong.
type RouteCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

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

// Create creates a Route in Kong. TODO Doc
func (s *RouteCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	route, current, _, client := argsFroRoute(arg...)
	// find the service to associate this route with
	svc, err := current.GetService(*route.Service.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find service associated with route %+v", route)
	}
	route.Service = svc.Service.DeepCopy()
	createdService, err := client.Routes.Create(nil, &route.Route)
	if err != nil {
		return nil, err
	}
	return createdService, nil
}

// Delete deletes a Route in Kong. TODO Doc
func (s *RouteCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	route, _, _, client := argsFroRoute(arg...)
	err := client.Routes.Delete(nil, route.ID)
	return nil, err
}

// Update updates a Route in Kong. TODO Doc
func (s *RouteCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	route, _, _, client := argsFroRoute(arg...)
	updatedService, err := client.Routes.Update(nil, &route.Route)
	if err != nil {
		return nil, err
	}
	return updatedService, nil
}
