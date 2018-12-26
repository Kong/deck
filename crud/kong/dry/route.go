package dry

import (
	"github.com/kong/deck/crud"
	arg "github.com/kong/deck/crud/kong"
	"github.com/kong/deck/print"
	"github.com/kong/deck/state"
)

// RouteCRUD implements Actions interface
// from the github.com/kong/crud package for the Route entitiy of Kong.
type RouteCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func routeFromStuct(arg arg.ArgStruct) *state.Route {
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
	print.CreatePrintln("creating route ", route)
	return nil, nil
}

// Delete deletes a Route in Kong. TODO Doc
func (s *RouteCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	route := routeFromStuct(argStruct)
	print.DeletePrintln("deleting route ", route)
	return nil, nil
}

// Update updates a Route in Kong. TODO Doc
func (s *RouteCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	route := routeFromStuct(argStruct)
	oldRouteObj, ok := argStruct.OldObj.(*state.Service)
	if !ok {
		panic("unexpected type, expected *state.service")
	}
	oldRoute := oldRouteObj.DeepCopy()
	// TODO remove this hack
	oldRoute.CreatedAt = nil
	oldRoute.UpdatedAt = nil
	diff := getDiff(oldRoute, route.DeepCopy())
	print.UpdatePrintln("updating route", route)
	print.UpdatePrintf("%s", diff)
	return nil, nil
}
