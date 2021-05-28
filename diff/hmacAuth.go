package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteHMACAuths() error {
	currentHMACAuths, err := sc.currentState.HMACAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching hmac-auths from state: %w", err)
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

func (sc *Syncer) deleteHMACAuth(hmacAuth *state.HMACAuth) (*crud.Event, error) {
	_, err := sc.targetState.HMACAuths.Get(*hmacAuth.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "hmac-auth",
			Obj:  hmacAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up hmac-auth %q: %w",
			*hmacAuth.Username, err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateHMACAuths() error {
	targetHMACAuths, err := sc.targetState.HMACAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching hmac-auths from state: %w", err)
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

func (sc *Syncer) createUpdateHMACAuth(hmacAuth *state.HMACAuth) (*crud.Event, error) {
	hmacAuth = &state.HMACAuth{HMACAuth: *hmacAuth.DeepCopy()}
	currentHMACAuth, err := sc.currentState.HMACAuths.Get(*hmacAuth.ID)
	if err == state.ErrNotFound {
		// hmacAuth not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: "hmac-auth",
			Obj:  hmacAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up hmac-auth %q: %w",
			*hmacAuth.Username, err)
	}
	// found, check if update needed

	if !currentHMACAuth.EqualWithOpts(hmacAuth, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "hmac-auth",
			Obj:    hmacAuth,
			OldObj: currentHMACAuth,
		}, nil
	}
	return nil, nil
}
