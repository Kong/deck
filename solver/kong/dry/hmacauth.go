package dry

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
)

// HMACAuthCRUD implements Actions interface
// from the github.com/kong/crud package for the hmac-auth of Kong.
type HMACAuthCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func hmacAuthFromStruct(arg diff.Event) *state.HMACAuth {
	hmacAuth, ok := arg.Obj.(*state.HMACAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return hmacAuth
}

// Create creates a fake hmac-auth.
// The arg should be of type diff.Event, containing the hmacAuth to be created,
// else the function will panic.
// It returns a the created *state.HMACAuth.
func (s *HMACAuthCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	hmacAuth := hmacAuthFromStruct(event)
	print.CreatePrintln("creating hmac-auth with username ", *hmacAuth.Username,
		" for consumer", *hmacAuth.Consumer.ID)
	return hmacAuth, nil
}

// Delete deletes a fake Route.
// The arg should be of type diff.Event, containing the hmacAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *HMACAuthCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	hmacAuth := hmacAuthFromStruct(event)
	print.DeletePrintln("deleting hmac-auth with username ", *hmacAuth.Username,
		" for consumer", *hmacAuth.Consumer.ID)
	return hmacAuth, nil
}

// Update updates a fake Route.
// The arg should be of type diff.Event, containing the hmacAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *HMACAuthCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	hmacAuth := hmacAuthFromStruct(event)
	oldRoute, ok := event.OldObj.(*state.HMACAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}
	// TODO remove this hack
	oldRoute.CreatedAt = nil

	diffString, err := getDiff(oldRoute.HMACAuth, hmacAuth.HMACAuth)
	if err != nil {
		return nil, err
	}
	print.UpdatePrintf("updating hmac-auth %s\n%s", *hmacAuth.Username, diffString)
	return hmacAuth, nil
}
