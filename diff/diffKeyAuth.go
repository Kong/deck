package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
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
	if keyAuth.Consumer == nil ||
		(utils.Empty(keyAuth.Consumer.ID)) {
		return nil, errors.Errorf("key-auth has no associated consumer: %+v",
			keyAuth.Key)
	}
	// lookup by Name
	_, err := sc.targetState.KeyAuths.Get(*keyAuth.Key)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "key-auth",
			Obj:  keyAuth,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up key-auth '%v'", *keyAuth.Key)
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
	currentKeyAuth, err := sc.currentState.KeyAuths.Get(*keyAuth.Key)
	if err == state.ErrNotFound {
		// keyAuth not present, create it

		// XXX fill foreign
		consumer, err := sc.currentState.Consumers.Get(*keyAuth.Consumer.Username)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not find consumer '%v' for key-auth %+v",
				*keyAuth.Consumer.Username, *keyAuth.Key)
		}
		keyAuth.Consumer = &consumer.Consumer
		// XXX

		keyAuth.ID = nil
		return &Event{
			Op:   crud.Create,
			Kind: "key-auth",
			Obj:  keyAuth,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up key-auth %v",
			*keyAuth.Key)
	}
	currentKeyAuth = &state.KeyAuth{KeyAuth: *currentKeyAuth.DeepCopy()}
	// found, check if update needed

	currentKeyAuth.Consumer = &kong.Consumer{
		Username: currentKeyAuth.Consumer.Username,
	}
	keyAuth.Consumer = &kong.Consumer{Username: keyAuth.Consumer.Username}
	if !currentKeyAuth.EqualWithOpts(keyAuth, true, true, false) {
		keyAuth.ID = kong.String(*currentKeyAuth.ID)

		// XXX fill foreign
		consumer, err := sc.currentState.Consumers.Get(*keyAuth.Consumer.Username)
		if err != nil {
			return nil, errors.Wrapf(err,
				"looking up service '%v' for key-auth '%v'",
				*keyAuth.Consumer.Username, *keyAuth.Key)
		}
		keyAuth.Consumer.ID = consumer.ID
		// XXX
		return &Event{
			Op:     crud.Update,
			Kind:   "key-auth",
			Obj:    keyAuth,
			OldObj: currentKeyAuth,
		}, nil
	}
	return nil, nil
}
