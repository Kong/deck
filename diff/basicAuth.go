package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
	"github.com/pkg/errors"
)

const (
	basicAuthPasswordWarning = "Warning: please note that changes in " +
		"password of basic-auth credentials are not detected by decK!!"
)

func (sc *Syncer) warnBasicAuth() {
	sc.once.Do(func() {
		if sc.SilenceWarnings {
			return
		}
		print.UpdatePrintln(basicAuthPasswordWarning)
	})
}

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
		return nil, errors.Wrapf(err, "error looking up basic-auth %v",
			*basicAuth.Username)
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
