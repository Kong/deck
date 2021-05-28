package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteMTLSAuths() error {
	currentMTLSAuths, err := sc.currentState.MTLSAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching mtls-auths from state: %w", err)
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

func (sc *Syncer) deleteMTLSAuth(mtlsAuth *state.MTLSAuth) (*crud.Event, error) {
	_, err := sc.targetState.MTLSAuths.Get(*mtlsAuth.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "mtls-auth",
			Obj:  mtlsAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up mtls-auth %q: %w", *mtlsAuth.ID, err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateMTLSAuths() error {
	targetMTLSAuths, err := sc.targetState.MTLSAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching mtls-auths from state: %w", err)
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

func (sc *Syncer) createUpdateMTLSAuth(mtlsAuth *state.MTLSAuth) (*crud.Event, error) {
	mtlsAuth = &state.MTLSAuth{MTLSAuth: *mtlsAuth.DeepCopy()}
	currentMTLSAuth, err := sc.currentState.MTLSAuths.Get(*mtlsAuth.ID)
	if err == state.ErrNotFound {
		// mtlsAuth not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: "mtls-auth",
			Obj:  mtlsAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up mtls-auth %q: %w",
			*mtlsAuth.ID, err)
	}
	// found, check if update needed

	if !currentMTLSAuth.EqualWithOpts(mtlsAuth, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "mtls-auth",
			Obj:    mtlsAuth,
			OldObj: currentMTLSAuth,
		}, nil
	}
	return nil, nil
}
