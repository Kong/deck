// Package dry contains Action for Kong entites.
// The actions are fake, meaning, the operations
// don't actually make REST Calls to Kong but
// mimic them.
package dry

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
)

// RouteCRUD implements Actions interface
// from the github.com/kong/crud package for the Route entitiy of Kong.
type RouteCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func routeFromStuct(arg diff.Event) *state.Route {
	route, ok := arg.Obj.(*state.Route)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return route
}

// Create creates a fake Route.
// The arg should be of type diff.Event, containing the route to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *RouteCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	route := routeFromStuct(event)
	print.CreatePrintln("creating route", *route.Name)
	route.ID = kong.String(utils.UUID())
	return route, nil
}

// Delete deletes a fake Route.
// The arg should be of type diff.Event, containing the route to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *RouteCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	route := routeFromStuct(event)
	print.DeletePrintln("deleting route", *route.Name)
	return route, nil
}

// Update updates a fake Route.
// The arg should be of type diff.Event, containing the route to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *RouteCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	route := routeFromStuct(event)
	oldRoute, ok := event.OldObj.(*state.Route)
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

	diffString, err := getDiff(oldRoute.Route, route.Route)
	if err != nil {
		return nil, err
	}
	print.UpdatePrintf("updating route %s\n%s", *route.Name, diffString)
	return route, nil
}
