package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteBasicAuths() error {
	currentBasicAuths, err := sc.currentState.BasicAuths.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching basic-auths from state")
	}

	for _, basicAuth := range currentBasicAuths {
		n, err := sc.deleteBasicAuth(basicAuth)
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

func (sc *Syncer) deleteBasicAuth(basicAuth *state.BasicAuth) (*Event, error) {
	if basicAuth.Consumer == nil ||
		(utils.Empty(basicAuth.Consumer.ID)) {
		return nil, errors.Errorf("basic-auth has no associated consumer: %+v",
			*basicAuth.Username)
	}
	// lookup by Name
	_, err := sc.targetState.BasicAuths.Get(*basicAuth.Username)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "basic-auth",
			Obj:  basicAuth,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up basic-auth '%v'",
			*basicAuth.Username)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateBasicAuths() error {
	targetBasicAuths, err := sc.targetState.BasicAuths.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching basic-auths from state")
	}

	for _, basicAuth := range targetBasicAuths {
		n, err := sc.createUpdateBasicAuth(basicAuth)
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

func (sc *Syncer) createUpdateBasicAuth(basicAuth *state.BasicAuth) (*Event, error) {
	basicAuth = &state.BasicAuth{BasicAuth: *basicAuth.DeepCopy()}
	currentBasicAuth, err := sc.currentState.BasicAuths.Get(*basicAuth.Username)
	if err == state.ErrNotFound {
		// basicAuth not present, create it

		// XXX fill foreign
		consumer, err := sc.currentState.Consumers.Get(*basicAuth.Consumer.Username)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not find consumer '%v' for basic-auth %+v",
				*basicAuth.Consumer.Username, *basicAuth.Username)
		}
		basicAuth.Consumer = &consumer.Consumer
		// XXX

		basicAuth.ID = nil
		return &Event{
			Op:   crud.Create,
			Kind: "basic-auth",
			Obj:  basicAuth,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up basic-auth %v",
			*basicAuth.Username)
	}
	currentBasicAuth = &state.BasicAuth{BasicAuth: *currentBasicAuth.DeepCopy()}
	// found, check if update needed

	currentBasicAuth.Consumer = &kong.Consumer{
		Username: currentBasicAuth.Consumer.Username,
	}
	basicAuth.Consumer = &kong.Consumer{Username: basicAuth.Consumer.Username}
	if !currentBasicAuth.EqualWithOpts(basicAuth, true, true, true, false) {
		basicAuth.ID = kong.String(*currentBasicAuth.ID)

		// XXX fill foreign
		consumer, err := sc.currentState.Consumers.Get(*basicAuth.Consumer.Username)
		if err != nil {
			return nil, errors.Wrapf(err,
				"looking up service '%v' for basic-auth '%v'",
				*basicAuth.Consumer.Username, *basicAuth.Username)
		}
		basicAuth.Consumer.ID = consumer.ID
		// XXX
		return &Event{
			Op:     crud.Update,
			Kind:   "basic-auth",
			Obj:    basicAuth,
			OldObj: currentBasicAuth,
		}, nil
	}
	return nil, nil
}
