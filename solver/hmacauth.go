package solver

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

// hmacAuthCRUD implements crud.Actions interface.
type hmacAuthCRUD struct {
	client *kong.Client
}

func hmacAuthFromStruct(arg diff.Event) *state.HMACAuth {
	hmacAuth, ok := arg.Obj.(*state.HMACAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return hmacAuth
}

// Create creates a Route in Kong.
// The arg should be of type diff.Event, containing the hmacAuth to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *hmacAuthCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	hmacAuth := hmacAuthFromStruct(event)
	cid := ""
	if !utils.Empty(hmacAuth.Consumer.Username) {
		cid = *hmacAuth.Consumer.Username
	}
	if !utils.Empty(hmacAuth.Consumer.ID) {
		cid = *hmacAuth.Consumer.ID
	}
	createdHMACAuth, err := s.client.HMACAuths.Create(nil, &cid,
		&hmacAuth.HMACAuth)
	if err != nil {
		return nil, err
	}
	return &state.HMACAuth{HMACAuth: *createdHMACAuth}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type diff.Event, containing the hmacAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *hmacAuthCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	hmacAuth := hmacAuthFromStruct(event)
	cid := ""
	if !utils.Empty(hmacAuth.Consumer.Username) {
		cid = *hmacAuth.Consumer.Username
	}
	if !utils.Empty(hmacAuth.Consumer.ID) {
		cid = *hmacAuth.Consumer.ID
	}
	err := s.client.HMACAuths.Delete(nil, &cid, hmacAuth.ID)
	if err != nil {
		return nil, err
	}
	return hmacAuth, nil
}

// Update updates a Route in Kong.
// The arg should be of type diff.Event, containing the hmacAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *hmacAuthCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	hmacAuth := hmacAuthFromStruct(event)

	cid := ""
	if !utils.Empty(hmacAuth.Consumer.Username) {
		cid = *hmacAuth.Consumer.Username
	}
	if !utils.Empty(hmacAuth.Consumer.ID) {
		cid = *hmacAuth.Consumer.ID
	}
	updatedHMACAuth, err := s.client.HMACAuths.Create(nil, &cid, &hmacAuth.HMACAuth)
	if err != nil {
		return nil, err
	}
	return &state.HMACAuth{HMACAuth: *updatedHMACAuth}, nil
}
