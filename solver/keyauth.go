package solver

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
)

// keyAuthCRUD implements crud.Actions interface.
type keyAuthCRUD struct {
	client *kong.Client
}

func keyAuthFromStuct(arg diff.Event) *state.KeyAuth {
	keyAuth, ok := arg.Obj.(*state.KeyAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return keyAuth
}

// Create creates a Route in Kong.
// The arg should be of type diff.Event, containing the keyAuth to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *keyAuthCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	keyAuth := keyAuthFromStuct(event)
	createdKeyAuth, err := s.client.KeyAuths.Create(nil, keyAuth.Consumer.ID,
		&keyAuth.KeyAuth)
	if err != nil {
		return nil, err
	}
	return &state.KeyAuth{KeyAuth: *createdKeyAuth}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type diff.Event, containing the keyAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *keyAuthCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	keyAuth := keyAuthFromStuct(event)
	cid := ""
	if !utils.Empty(keyAuth.Consumer.Username) {
		cid = *keyAuth.Consumer.Username
	}
	if !utils.Empty(keyAuth.Consumer.ID) {
		cid = *keyAuth.Consumer.ID
	}
	err := s.client.KeyAuths.Delete(nil, &cid, keyAuth.ID)
	if err != nil {
		return nil, err
	}
	return keyAuth, nil
}

// Update updates a Route in Kong.
// The arg should be of type diff.Event, containing the keyAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *keyAuthCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	keyAuth := keyAuthFromStuct(event)

	updatedKeyAuth, err := s.client.KeyAuths.Create(nil, keyAuth.Consumer.ID,
		&keyAuth.KeyAuth)
	if err != nil {
		return nil, err
	}
	return &state.KeyAuth{KeyAuth: *updatedKeyAuth}, nil
}
