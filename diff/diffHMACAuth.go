package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
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
	if hmacAuth.Consumer == nil ||
		(utils.Empty(hmacAuth.Consumer.ID)) {
		return nil, errors.Errorf("hmac-auth has no associated consumer: %+v",
			*hmacAuth.Username)
	}
	// lookup by Name
	_, err := sc.targetState.HMACAuths.Get(*hmacAuth.Username)
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
	currentHMACAuth, err := sc.currentState.HMACAuths.Get(*hmacAuth.Username)
	if err == state.ErrNotFound {
		// hmacAuth not present, create it

		// XXX fill foreign
		consumer, err := sc.currentState.Consumers.Get(*hmacAuth.Consumer.Username)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not find consumer '%v' for hmac-auth %+v",
				*hmacAuth.Consumer.Username, *hmacAuth.Username)
		}
		hmacAuth.Consumer = &consumer.Consumer
		// XXX

		hmacAuth.ID = nil
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
	currentHMACAuth = &state.HMACAuth{HMACAuth: *currentHMACAuth.DeepCopy()}
	// found, check if update needed

	currentHMACAuth.Consumer = &kong.Consumer{
		Username: currentHMACAuth.Consumer.Username,
	}
	hmacAuth.Consumer = &kong.Consumer{Username: hmacAuth.Consumer.Username}
	if !currentHMACAuth.EqualWithOpts(hmacAuth, true, true, false) {
		hmacAuth.ID = kong.String(*currentHMACAuth.ID)

		// XXX fill foreign
		consumer, err := sc.currentState.Consumers.Get(*hmacAuth.Consumer.Username)
		if err != nil {
			return nil, errors.Wrapf(err,
				"looking up service '%v' for hmac-auth '%v'",
				*hmacAuth.Consumer.Username, *hmacAuth.Username)
		}
		hmacAuth.Consumer.ID = consumer.ID
		// XXX
		return &Event{
			Op:     crud.Update,
			Kind:   "hmac-auth",
			Obj:    hmacAuth,
			OldObj: currentHMACAuth,
		}, nil
	}
	return nil, nil
}
