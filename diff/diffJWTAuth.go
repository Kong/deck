package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
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
	if jwtAuth.Consumer == nil ||
		(utils.Empty(jwtAuth.Consumer.ID)) {
		return nil, errors.Errorf("jwt-auth has no associated consumer: %+v",
			*jwtAuth.Key)
	}
	// lookup by Name
	_, err := sc.targetState.JWTAuths.Get(*jwtAuth.Key)
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
	currentJWTAuth, err := sc.currentState.JWTAuths.Get(*jwtAuth.Key)
	if err == state.ErrNotFound {
		// jwtAuth not present, create it

		// XXX fill foreign
		consumer, err := sc.currentState.Consumers.Get(*jwtAuth.Consumer.Username)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not find consumer '%v' for jwt-auth %+v",
				*jwtAuth.Consumer.Username, *jwtAuth.Key)
		}
		jwtAuth.Consumer = &consumer.Consumer
		// XXX

		jwtAuth.ID = nil
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
	currentJWTAuth = &state.JWTAuth{JWTAuth: *currentJWTAuth.DeepCopy()}
	// found, check if update needed

	currentJWTAuth.Consumer = &kong.Consumer{
		Username: currentJWTAuth.Consumer.Username,
	}
	jwtAuth.Consumer = &kong.Consumer{Username: jwtAuth.Consumer.Username}
	if !currentJWTAuth.EqualWithOpts(jwtAuth, true, true, false) {
		jwtAuth.ID = kong.String(*currentJWTAuth.ID)

		// XXX fill foreign
		consumer, err := sc.currentState.Consumers.Get(*jwtAuth.Consumer.Username)
		if err != nil {
			return nil, errors.Wrapf(err,
				"looking up service '%v' for jwt-auth '%v'",
				*jwtAuth.Consumer.Username, *jwtAuth.Key)
		}
		jwtAuth.Consumer.ID = consumer.ID
		// XXX
		return &Event{
			Op:     crud.Update,
			Kind:   "jwt-auth",
			Obj:    jwtAuth,
			OldObj: currentJWTAuth,
		}, nil
	}
	return nil, nil
}
