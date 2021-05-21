package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteHMACAuths() error {
	currentHMACAuths, err := sc.currentState.HMACAuths.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching hmac-auths from state")
	}

	for _, hmacAuth := range currentHMACAuths {
		n, err := sc.deleteHMACAuth(hmacAuth)
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

func (sc *Syncer) deleteHMACAuth(hmacAuth *state.HMACAuth) (*Event, error) {
	_, err := sc.targetState.HMACAuths.Get(*hmacAuth.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "hmac-auth",
			Obj:  hmacAuth,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up hmac-auth '%v'",
			*hmacAuth.Username)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateHMACAuths() error {
	targetHMACAuths, err := sc.targetState.HMACAuths.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching hmac-auths from state")
	}

	for _, hmacAuth := range targetHMACAuths {
		n, err := sc.createUpdateHMACAuth(hmacAuth)
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

func (sc *Syncer) createUpdateHMACAuth(hmacAuth *state.HMACAuth) (*Event, error) {
	hmacAuth = &state.HMACAuth{HMACAuth: *hmacAuth.DeepCopy()}
	currentHMACAuth, err := sc.currentState.HMACAuths.Get(*hmacAuth.ID)
	if err == state.ErrNotFound {
		// hmacAuth not present, create it

		return &Event{
			Op:   crud.Create,
			Kind: "hmac-auth",
			Obj:  hmacAuth,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up hmac-auth %v",
			*hmacAuth.Username)
	}
	// found, check if update needed

	if !currentHMACAuth.EqualWithOpts(hmacAuth, false, true, false) {
		return &Event{
			Op:     crud.Update,
			Kind:   "hmac-auth",
			Obj:    hmacAuth,
			OldObj: currentHMACAuth,
		}, nil
	}
	return nil, nil
}
