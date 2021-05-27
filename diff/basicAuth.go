package diff

import (
	"fmt"

	"github.com/kong/deck/cprint"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

const (
	basicAuthPasswordWarning = "Warning: import/export of basic-auth" +
		"credentials using decK doesn't work due to hashing of passwords in Kong."
)

func (sc *Syncer) warnBasicAuth() {
	sc.once.Do(func() {
		if sc.SilenceWarnings {
			return
		}
		cprint.UpdatePrintln(basicAuthPasswordWarning)
	})
}

func (sc *Syncer) deleteBasicAuths() error {
	currentBasicAuths, err := sc.currentState.BasicAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching basic-auths from state: %w", err)
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
	sc.warnBasicAuth()
	_, err := sc.targetState.BasicAuths.Get(*basicAuth.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "basic-auth",
			Obj:  basicAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up basic-auth %q: %w",
			*basicAuth.Username, err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateBasicAuths() error {
	targetBasicAuths, err := sc.targetState.BasicAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching basic-auths from state: %w", err)
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
	sc.warnBasicAuth()
	basicAuth = &state.BasicAuth{BasicAuth: *basicAuth.DeepCopy()}
	currentBasicAuth, err := sc.currentState.BasicAuths.Get(*basicAuth.ID)
	if err == state.ErrNotFound {
		// basicAuth not present, create it

		return &Event{
			Op:   crud.Create,
			Kind: "basic-auth",
			Obj:  basicAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up basic-auth %q: %w",
			*basicAuth.Username, err)
	}
	// found, check if update needed

	if !currentBasicAuth.EqualWithOpts(basicAuth, false, true, true, false) {
		return &Event{
			Op:     crud.Update,
			Kind:   "basic-auth",
			Obj:    basicAuth,
			OldObj: currentBasicAuth,
		}, nil
	}
	return nil, nil
}
