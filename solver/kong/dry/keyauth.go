package dry

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
)

// KeyAuthCRUD implements Actions interface
// from the github.com/kong/crud package for the key-auth of Kong.
type KeyAuthCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func keyAuthFromStruct(arg diff.Event) *state.KeyAuth {
	keyAuth, ok := arg.Obj.(*state.KeyAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return keyAuth
}

// Create creates a fake key-auth.
// The arg should be of type diff.Event, containing the keyAuth to be created,
// else the function will panic.
// It returns a the created *state.KeyAuth.
func (s *KeyAuthCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	keyAuth := keyAuthFromStruct(event)
	print.CreatePrintln("creating key-auth", stripKey(*keyAuth.Key),
		"(last 5) for consumer", *keyAuth.Consumer.ID)
	return keyAuth, nil
}

// Delete deletes a fake Route.
// The arg should be of type diff.Event, containing the keyAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *KeyAuthCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	keyAuth := keyAuthFromStruct(event)
	print.DeletePrintln("deleting key-auth", stripKey(*keyAuth.Key),
		"(last 5) for consumer", *keyAuth.Consumer.ID)
	return keyAuth, nil
}

// Update updates a fake Route.
// The arg should be of type diff.Event, containing the keyAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *KeyAuthCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	keyAuth := keyAuthFromStruct(event)
	oldRoute, ok := event.OldObj.(*state.KeyAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}
	// TODO remove this hack
	oldRoute.CreatedAt = nil

	diffString, err := getDiff(oldRoute.KeyAuth, keyAuth.KeyAuth)
	if err != nil {
		return nil, err
	}
	// TODO strip out or trim key diff
	// decK's logs will be sent to logging system as it is likely
	// users are using decK using a CI.
	// A diff of KeyAuth is unlikely but when this happens, we don't
	// want to be logging the entire key.
	// This is a larger problem with decK.
	print.UpdatePrintf("updating key-auth %s\n%s", *keyAuth.Key, diffString)
	return keyAuth, nil
}

// stripKey returns the last 5 characters of key.
// If key is less than or equal to 5 characters, then the key is returned as is.
func stripKey(key string) string {
	if len(key) <= 5 {
		return key
	}
	return string(key[len(key)-5:])
}
