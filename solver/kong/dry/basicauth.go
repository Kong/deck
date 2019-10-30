package dry

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
)

// BasicAuthCRUD implements Actions interface
// from the github.com/kong/crud package for the basic-auth of Kong.
type BasicAuthCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func basicAuthFromStruct(arg diff.Event) *state.BasicAuth {
	basicAuth, ok := arg.Obj.(*state.BasicAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return basicAuth
}

// Create creates a fake basic-auth.
// The arg should be of type diff.Event, containing the basicAuth to be created,
// else the function will panic.
// It returns a the created *state.BasicAuth.
func (s *BasicAuthCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	basicAuth := basicAuthFromStruct(event)
	print.CreatePrintln("creating basic-auth with username ", *basicAuth.Username,
		" for consumer", *basicAuth.Consumer.ID)
	return basicAuth, nil
}

// Delete deletes a fake Route.
// The arg should be of type diff.Event, containing the basicAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *BasicAuthCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	basicAuth := basicAuthFromStruct(event)
	print.DeletePrintln("deleting basic-auth with username ", *basicAuth.Username,
		" for consumer", *basicAuth.Consumer.ID)
	return basicAuth, nil
}

// Update updates a fake Route.
// The arg should be of type diff.Event, containing the basicAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *BasicAuthCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	basicAuth := basicAuthFromStruct(event)
	oldRoute, ok := event.OldObj.(*state.BasicAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}
	// TODO remove this hack
	oldRoute.CreatedAt = nil

	diffString, err := getDiff(oldRoute.BasicAuth, basicAuth.BasicAuth)
	if err != nil {
		return nil, err
	}
	print.UpdatePrintf("updating basic-auth %s\n%s", *basicAuth.Username,
		diffString)
	return basicAuth, nil
}
