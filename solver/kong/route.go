package kong

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
)

// RouteCRUD implements Actions interface
// from the github.com/kong/crud package for the Route entitiy of Kong.
type RouteCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func argStructFromArg(arg crud.Arg) diff.ArgStruct {
	argStruct, ok := arg.(diff.ArgStruct)
	if !ok {
		panic("unexpected type, expected ArgStruct")
	}
	return argStruct
}

func routeFromStuct(arg diff.ArgStruct) *state.Route {
	route, ok := arg.Obj.(*state.Route)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return route
}

// Create creates a Route in Kong. TODO Doc
func (s *RouteCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	route := routeFromStuct(argStruct)
	createdRoute, err := argStruct.Client.Routes.Create(nil, &route.Route)
	if err != nil {
		return nil, err
	}
	return &state.Route{Route: *createdRoute}, nil
}

// Delete deletes a Route in Kong. TODO Doc
func (s *RouteCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	route := routeFromStuct(argStruct)
	err := argStruct.Client.Routes.Delete(nil, route.ID)
	if err != nil {
		return nil, err
	}
	return route, nil
}

// Update updates a Route in Kong. TODO Doc
func (s *RouteCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	route := routeFromStuct(argStruct)

	updatedRoute, err := argStruct.Client.Routes.Update(nil, &route.Route)
	if err != nil {
		return nil, err
	}
	return &state.Route{Route: *updatedRoute}, nil
}
