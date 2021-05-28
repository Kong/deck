package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteJWTAuths() error {
	currentJWTAuths, err := sc.currentState.JWTAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching jwt-auths from state: %w", err)
	}

	for _, jwtAuth := range currentJWTAuths {
		n, err := sc.deleteJWTAuth(jwtAuth)
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

func (sc *Syncer) deleteJWTAuth(jwtAuth *state.JWTAuth) (*crud.Event, error) {
	_, err := sc.targetState.JWTAuths.Get(*jwtAuth.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "jwt-auth",
			Obj:  jwtAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up jwt-auth %q: %w", *jwtAuth.Key, err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateJWTAuths() error {
	targetJWTAuths, err := sc.targetState.JWTAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching jwt-auths from state: %w", err)
	}

	for _, jwtAuth := range targetJWTAuths {
		n, err := sc.createUpdateJWTAuth(jwtAuth)
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

func (sc *Syncer) createUpdateJWTAuth(jwtAuth *state.JWTAuth) (*crud.Event, error) {
	jwtAuth = &state.JWTAuth{JWTAuth: *jwtAuth.DeepCopy()}
	currentJWTAuth, err := sc.currentState.JWTAuths.Get(*jwtAuth.ID)
	if err == state.ErrNotFound {
		// jwtAuth not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: "jwt-auth",
			Obj:  jwtAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up jwt-auth %q: %w",
			*jwtAuth.Key, err)
	}
	// found, check if update needed

	if !currentJWTAuth.EqualWithOpts(jwtAuth, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "jwt-auth",
			Obj:    jwtAuth,
			OldObj: currentJWTAuth,
		}, nil
	}
	return nil, nil
}
