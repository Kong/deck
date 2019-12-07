package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteKeyAuths() error {
	currentKeyAuths, err := sc.currentState.KeyAuths.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching key-auths from state")
	}

	for _, keyAuth := range currentKeyAuths {
		n, err := sc.deleteKeyAuth(keyAuth)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) deleteKeyAuth(keyAuth *state.KeyAuth) (*Event, error) {
	_, err := sc.targetState.KeyAuths.Get(*keyAuth.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "key-auth",
			Obj:  keyAuth,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up key-auth '%v'", *keyAuth.ID)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateKeyAuths() error {
	targetKeyAuths, err := sc.targetState.KeyAuths.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching key-auths from state")
	}

	for _, keyAuth := range targetKeyAuths {
		n, err := sc.createUpdateKeyAuth(keyAuth)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) createUpdateKeyAuth(keyAuth *state.KeyAuth) (*Event, error) {
	keyAuth = &state.KeyAuth{KeyAuth: *keyAuth.DeepCopy()}
	currentKeyAuth, err := sc.currentState.KeyAuths.Get(*keyAuth.ID)
	if err == state.ErrNotFound {
		// keyAuth not present, create it

		return &Event{
			Op:   crud.Create,
			Kind: "key-auth",
			Obj:  keyAuth,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up key-auth %v",
			*keyAuth.ID)
	}
	// found, check if update needed

	if !currentKeyAuth.EqualWithOpts(keyAuth, false, true, false) {

		return &Event{
			Op:     crud.Update,
			Kind:   "key-auth",
			Obj:    keyAuth,
			OldObj: currentKeyAuth,
		}, nil
	}
	return nil, nil
}
