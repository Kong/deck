package types

import (
	"context"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

// keyAuthCRUD implements crud.Actions interface.
type keyAuthCRUD struct {
	client *kong.Client
}

func keyAuthFromStruct(arg crud.Event) *state.KeyAuth {
	keyAuth, ok := arg.Obj.(*state.KeyAuth)
	if !ok {
		panic("unexpected type, expected *state.KeyAuth")
	}

	return keyAuth
}

// Create creates a Route in Kong.
// The arg should be of type crud.Event, containing the keyAuth to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *keyAuthCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	keyAuth := keyAuthFromStruct(event)
	createdKeyAuth, err := s.client.KeyAuths.Create(ctx, keyAuth.Consumer.ID,
		&keyAuth.KeyAuth)
	if err != nil {
		return nil, err
	}
	return &state.KeyAuth{KeyAuth: *createdKeyAuth}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type crud.Event, containing the keyAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *keyAuthCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	keyAuth := keyAuthFromStruct(event)
	cid := ""
	if !utils.Empty(keyAuth.Consumer.Username) {
		cid = *keyAuth.Consumer.Username
	}
	if !utils.Empty(keyAuth.Consumer.ID) {
		cid = *keyAuth.Consumer.ID
	}
	err := s.client.KeyAuths.Delete(ctx, &cid, keyAuth.ID)
	if err != nil {
		return nil, err
	}
	return keyAuth, nil
}

// Update updates a Route in Kong.
// The arg should be of type crud.Event, containing the keyAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *keyAuthCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	keyAuth := keyAuthFromStruct(event)

	updatedKeyAuth, err := s.client.KeyAuths.Create(ctx, keyAuth.Consumer.ID,
		&keyAuth.KeyAuth)
	if err != nil {
		return nil, err
	}
	return &state.KeyAuth{KeyAuth: *updatedKeyAuth}, nil
}

type keyAuthDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *keyAuthDiffer) Deletes(handler func(crud.Event) error) error {
	currentKeyAuths, err := d.currentState.KeyAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching key-auths from state: %w", err)
	}

	for _, keyAuth := range currentKeyAuths {
		n, err := d.deleteKeyAuth(keyAuth)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *keyAuthDiffer) deleteKeyAuth(keyAuth *state.KeyAuth) (*crud.Event, error) {
	_, err := d.targetState.KeyAuths.Get(*keyAuth.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  keyAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up key-auth %q: %w", *keyAuth.ID, err)
	}
	return nil, nil
}

func (d *keyAuthDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetKeyAuths, err := d.targetState.KeyAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching key-auths from state: %w", err)
	}

	for _, keyAuth := range targetKeyAuths {
		n, err := d.createUpdateKeyAuth(keyAuth)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *keyAuthDiffer) createUpdateKeyAuth(keyAuth *state.KeyAuth) (*crud.Event, error) {
	keyAuth = &state.KeyAuth{KeyAuth: *keyAuth.DeepCopy()}
	currentKeyAuth, err := d.currentState.KeyAuths.Get(*keyAuth.ID)
	if err == state.ErrNotFound {
		// keyAuth not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  keyAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up key-auth %q: %w",
			*keyAuth.ID, err)
	}
	// found, check if update needed

	if !currentKeyAuth.EqualWithOpts(keyAuth, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    keyAuth,
			OldObj: currentKeyAuth,
		}, nil
	}
	return nil, nil
}
