package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteOauth2Creds() error {
	currentOauth2Creds, err := sc.currentState.Oauth2Creds.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching oauth2-cred from state: %w", err)
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
	*crud.Event, error) {
	_, err := sc.targetState.Oauth2Creds.Get(*oauth2Cred.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "oauth2-cred",
			Obj:  oauth2Cred,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up oauth2-cred %q: %w", *oauth2Cred.Name, err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateOauth2Creds() error {
	targetOauth2Creds, err := sc.targetState.Oauth2Creds.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching oauth2-creds from state: %w", err)
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

func (sc *Syncer) createUpdateOauth2Cred(oauth2Cred *state.Oauth2Credential) (*crud.Event, error) {
	oauth2Cred = &state.Oauth2Credential{Oauth2Credential: *oauth2Cred.DeepCopy()}
	currentOauth2Cred, err := sc.currentState.Oauth2Creds.Get(*oauth2Cred.ID)
	if err == state.ErrNotFound {
		// oauth2Cred not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: "oauth2-cred",
			Obj:  oauth2Cred,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up oauth2-cred %q: %w",
			*oauth2Cred.Name, err)
	}
	currentOauth2Cred = &state.Oauth2Credential{Oauth2Credential: *currentOauth2Cred.DeepCopy()}
	// found, check if update needed

	if !currentOauth2Cred.EqualWithOpts(oauth2Cred, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "oauth2-cred",
			Obj:    oauth2Cred,
			OldObj: currentOauth2Cred,
		}, nil
	}
	return nil, nil
}
