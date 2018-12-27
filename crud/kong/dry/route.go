package dry

import (
	"github.com/hbagdi/go-kong/kong"
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
	print.CreatePrintln("creating route ", *route.Name)
	return nil, nil
}

// Delete deletes a Route in Kong. TODO Doc
func (s *RouteCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	route := routeFromStuct(argStruct)
	print.DeletePrintln("deleting route ", *route.Name)
	return nil, nil
}

// Update updates a Route in Kong. TODO Doc
func (s *RouteCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	route := routeFromStuct(argStruct)
	oldRoute, ok := argStruct.OldObj.(*state.Route)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}
	// TODO remove this hack
	oldRoute.CreatedAt = nil
	oldRoute.UpdatedAt = nil
	oldRoute.Service = &kong.Service{Name: oldRoute.Service.Name}
	oldRoute.ID = nil

	route.ID = nil
	route.Service = &kong.Service{Name: route.Service.Name}

	diff := getDiff(oldRoute.Route, route.Route)
	print.UpdatePrintln("updating route", *route.Name)
	print.UpdatePrintf("%s", diff)
	return nil, nil
}
