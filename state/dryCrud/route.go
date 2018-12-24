package drycrud

import (
	"fmt"

	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
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

	return route, nil, nil, nil
}

// Create creates a Route in Kong. TODO Doc
func (s *RouteCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	route, _, _, _ := argsFroRoute(arg...)
	fmt.Println("creating route", route)
	return nil, nil
}

// Delete deletes a Route in Kong. TODO Doc
func (s *RouteCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	route, _, _, _ := argsFroRoute(arg...)
	fmt.Println("deleting route", route)
	return nil, nil
}

// Update updates a Route in Kong. TODO Doc
func (s *RouteCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	route, _, _, _ := argsFroRoute(arg...)
	fmt.Println("updating route", route)
	return nil, nil
}
