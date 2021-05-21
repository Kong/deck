package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
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

func (sc *Syncer) deleteOauth2Cred(oauth2Cred *state.Oauth2Credential) (
	*Event, error) {
	_, err := sc.targetState.Oauth2Creds.Get(*oauth2Cred.ID)
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
	currentOauth2Cred, err := sc.currentState.Oauth2Creds.Get(*oauth2Cred.ID)
	if err == state.ErrNotFound {
		// oauth2Cred not present, create it

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

	if !currentOauth2Cred.EqualWithOpts(oauth2Cred, false, true, false) {
		return &Event{
			Op:     crud.Update,
			Kind:   "oauth2-cred",
			Obj:    oauth2Cred,
			OldObj: currentOauth2Cred,
		}, nil
	}
	return nil, nil
}
