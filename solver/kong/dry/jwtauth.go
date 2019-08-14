package dry

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
)

// JWTAuthCRUD implements Actions interface
// from the github.com/kong/crud package for the jwt-secret of Kong.
type JWTAuthCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func jwtAuthFromStruct(arg diff.Event) *state.JWTAuth {
	jwtAuth, ok := arg.Obj.(*state.JWTAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return jwtAuth
}

// Create creates a fake jwt-secret.
// The arg should be of type diff.Event, containing the jwtAuth to be created,
// else the function will panic.
// It returns a the created *state.JWTAuth.
func (s *JWTAuthCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	jwtAuth := jwtAuthFromStruct(event)
	print.CreatePrintln("creating jwt-secret", *jwtAuth.Key,
		"for consumer", *jwtAuth.Consumer.Username)
	jwtAuth.ID = kong.String(utils.UUID())
	return jwtAuth, nil
}

// Delete deletes a fake Route.
// The arg should be of type diff.Event, containing the jwtAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *JWTAuthCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	jwtAuth := jwtAuthFromStruct(event)
	print.DeletePrintln("deleting jwt-secret", *jwtAuth.Key,
		"for consumer", *jwtAuth.Consumer.Username)
	return jwtAuth, nil
}

// Update updates a fake Route.
// The arg should be of type diff.Event, containing the jwtAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *JWTAuthCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	jwtAuth := jwtAuthFromStruct(event)
	oldRoute, ok := event.OldObj.(*state.JWTAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}
	// TODO remove this hack
	oldRoute.CreatedAt = nil
	oldRoute.Consumer = &kong.Consumer{Username: oldRoute.Consumer.Username}
	oldRoute.ID = nil

	jwtAuth.ID = nil
	jwtAuth.Consumer = &kong.Consumer{Username: jwtAuth.Consumer.Username}

	diffString, err := getDiff(oldRoute.JWTAuth, jwtAuth.JWTAuth)
	if err != nil {
		return nil, err
	}
	print.UpdatePrintf("updating jwt-secret %s\n%s", *jwtAuth.Key, diffString)
	return jwtAuth, nil
}
