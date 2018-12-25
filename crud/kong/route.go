package kong

import (
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

func argStructFromArg(arg crud.Arg) ArgStruct {
	argStruct, ok := arg.(ArgStruct)
	if !ok {
		panic("unexpected type, expected ArgStruct")
	}
	return argStruct
}

func routeFromStuct(arg ArgStruct) *state.Route {
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

	// find the service to associate this route with
	svc, err := argStruct.CurrentState.GetService(*route.Service.Name)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to find service associated with route %+v", route)
	}
	route.Service = svc.Service.DeepCopy()
	createdService, err := argStruct.Client.Routes.Create(nil, &route.Route)
	if err != nil {
		return nil, err
	}
	return createdService, nil
}

// Delete deletes a Route in Kong. TODO Doc
func (s *RouteCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	route := routeFromStuct(argStruct)

	err := argStruct.Client.Routes.Delete(nil, route.ID)
	return nil, err
}

// Update updates a Route in Kong. TODO Doc
func (s *RouteCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	route := routeFromStuct(argStruct)

	updatedService, err := argStruct.Client.Routes.Update(nil, &route.Route)
	if err != nil {
		return nil, err
	}
	return updatedService, nil
}
