package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteOauth2Creds() error {
	currentOauth2Creds, err := sc.currentState.Oauth2Creds.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching oauth2-cred from state")
	}

	for _, oauth2Cred := range currentOauth2Creds {
		n, err := sc.deleteOauth2Cred(oauth2Cred)
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

func (sc *Syncer) deleteOauth2Cred(oauth2Cred *state.Oauth2Credential) (*Event, error) {
	if oauth2Cred.Consumer == nil ||
		(utils.Empty(oauth2Cred.Consumer.ID)) {
		return nil, errors.Errorf("oauth2-cred has no associated consumer: %+v",
			*oauth2Cred.Name)
	}
	// lookup by Name
	_, err := sc.targetState.Oauth2Creds.Get(*oauth2Cred.ClientID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "oauth2-cred",
			Obj:  oauth2Cred,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up oauth2-cred '%v'", *oauth2Cred.Name)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateOauth2Creds() error {
	targetOauth2Creds, err := sc.targetState.Oauth2Creds.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching oauth2-creds from state")
	}

	for _, oauth2Cred := range targetOauth2Creds {
		n, err := sc.createUpdateOauth2Cred(oauth2Cred)
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

func (sc *Syncer) createUpdateOauth2Cred(oauth2Cred *state.Oauth2Credential) (*Event, error) {
	oauth2Cred = &state.Oauth2Credential{Oauth2Credential: *oauth2Cred.DeepCopy()}
	currentOauth2Cred, err := sc.currentState.Oauth2Creds.Get(*oauth2Cred.ClientID)
	if err == state.ErrNotFound {
		// oauth2Cred not present, create it

		// XXX fill foreign
		consumer, err := sc.currentState.Consumers.Get(*oauth2Cred.Consumer.Username)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not find consumer '%v' for oauth2-cred %+v",
				*oauth2Cred.Consumer.Username, *oauth2Cred.Name)
		}
		oauth2Cred.Consumer = &consumer.Consumer
		// XXX

		oauth2Cred.ID = nil
		return &Event{
			Op:   crud.Create,
			Kind: "oauth2-cred",
			Obj:  oauth2Cred,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up oauth2-cred %v",
			*oauth2Cred.Name)
	}
	currentOauth2Cred = &state.Oauth2Credential{Oauth2Credential: *currentOauth2Cred.DeepCopy()}
	// found, check if update needed

	currentOauth2Cred.Consumer = &kong.Consumer{
		Username: currentOauth2Cred.Consumer.Username,
	}
	oauth2Cred.Consumer = &kong.Consumer{Username: oauth2Cred.Consumer.Username}
	if !currentOauth2Cred.EqualWithOpts(oauth2Cred, true, true, false) {
		oauth2Cred.ID = kong.String(*currentOauth2Cred.ID)

		// XXX fill foreign
		consumer, err := sc.currentState.Consumers.Get(*oauth2Cred.Consumer.Username)
		if err != nil {
			return nil, errors.Wrapf(err,
				"looking up service '%v' for oauth2-cred '%v'",
				*oauth2Cred.Consumer.Username, *oauth2Cred.Name)
		}
		oauth2Cred.Consumer.ID = consumer.ID
		// XXX
		return &Event{
			Op:     crud.Update,
			Kind:   "oauth2-cred",
			Obj:    oauth2Cred,
			OldObj: currentOauth2Cred,
		}, nil
	}
	return nil, nil
}
