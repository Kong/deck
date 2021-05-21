package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteMTLSAuths() error {
	currentMTLSAuths, err := sc.currentState.MTLSAuths.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching mtls-auths from state")
	}

	for _, mtlsAuth := range currentMTLSAuths {
		n, err := sc.deleteMTLSAuth(mtlsAuth)
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

func (sc *Syncer) deleteMTLSAuth(mtlsAuth *state.MTLSAuth) (*Event, error) {
	_, err := sc.targetState.MTLSAuths.Get(*mtlsAuth.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "mtls-auth",
			Obj:  mtlsAuth,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up mtls-auth '%v'", *mtlsAuth.ID)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateMTLSAuths() error {
	targetMTLSAuths, err := sc.targetState.MTLSAuths.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching mtls-auths from state")
	}

	for _, mtlsAuth := range targetMTLSAuths {
		n, err := sc.createUpdateMTLSAuth(mtlsAuth)
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

func (sc *Syncer) createUpdateMTLSAuth(mtlsAuth *state.MTLSAuth) (*Event, error) {
	mtlsAuth = &state.MTLSAuth{MTLSAuth: *mtlsAuth.DeepCopy()}
	currentMTLSAuth, err := sc.currentState.MTLSAuths.Get(*mtlsAuth.ID)
	if err == state.ErrNotFound {
		// mtlsAuth not present, create it

		return &Event{
			Op:   crud.Create,
			Kind: "mtls-auth",
			Obj:  mtlsAuth,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up mtls-auth %v",
			*mtlsAuth.ID)
	}
	// found, check if update needed

	if !currentMTLSAuth.EqualWithOpts(mtlsAuth, false, true, false) {
		return &Event{
			Op:     crud.Update,
			Kind:   "mtls-auth",
			Obj:    mtlsAuth,
			OldObj: currentMTLSAuth,
		}, nil
	}
	return nil, nil
}
