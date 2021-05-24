package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteJWTAuths() error {
	currentJWTAuths, err := sc.currentState.JWTAuths.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching jwt-auths from state")
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

func (sc *Syncer) deleteJWTAuth(jwtAuth *state.JWTAuth) (*Event, error) {
	_, err := sc.targetState.JWTAuths.Get(*jwtAuth.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "jwt-auth",
			Obj:  jwtAuth,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up jwt-auth '%v'", *jwtAuth.Key)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateJWTAuths() error {
	targetJWTAuths, err := sc.targetState.JWTAuths.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching jwt-auths from state")
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

func (sc *Syncer) createUpdateJWTAuth(jwtAuth *state.JWTAuth) (*Event, error) {
	jwtAuth = &state.JWTAuth{JWTAuth: *jwtAuth.DeepCopy()}
	currentJWTAuth, err := sc.currentState.JWTAuths.Get(*jwtAuth.ID)
	if err == state.ErrNotFound {
		// jwtAuth not present, create it

		return &Event{
			Op:   crud.Create,
			Kind: "jwt-auth",
			Obj:  jwtAuth,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up jwt-auth %v",
			*jwtAuth.Key)
	}
	// found, check if update needed

	if !currentJWTAuth.EqualWithOpts(jwtAuth, false, true, false) {
		return &Event{
			Op:     crud.Update,
			Kind:   "jwt-auth",
			Obj:    jwtAuth,
			OldObj: currentJWTAuth,
		}, nil
	}
	return nil, nil
}
