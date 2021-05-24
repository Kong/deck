package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteCACertificates() error {
	currentCACertificates, err := sc.currentState.CACertificates.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching caCertificates from state: %w", err)
	}

	for _, certificate := range currentCACertificates {
		n, err := sc.deleteCACertificate(certificate)
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

func (sc *Syncer) deleteCACertificate(
	caCert *state.CACertificate) (*Event, error) {
	_, err := sc.targetState.CACertificates.Get(*caCert.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "ca_certificate",
			Obj:  caCert,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up caCertificate %q: %w",
			caCert.Identifier(), err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateCACertificates() error {
	targetCACertificates, err := sc.targetState.CACertificates.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching caCertificates from state: %w", err)
	}

	for _, caCert := range targetCACertificates {
		n, err := sc.createUpdateCACertificate(caCert)
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

func (sc *Syncer) createUpdateCACertificate(
	caCert *state.CACertificate) (*Event, error) {
	caCertCopy := &state.CACertificate{CACertificate: *caCert.DeepCopy()}
	currentCACert, err := sc.currentState.CACertificates.Get(*caCert.ID)

	if err == state.ErrNotFound {
		// caCertificate not present, create it
		return &Event{
			Op:   crud.Create,
			Kind: "ca_certificate",
			Obj:  caCertCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up caCertificate %q: %w",
			caCert.Identifier(), err)
	}

	// found, check if update needed
	if !currentCACert.EqualWithOpts(caCertCopy, false, true) {
		return &Event{
			Op:     crud.Update,
			Kind:   "ca_certificate",
			Obj:    caCertCopy,
			OldObj: currentCACert,
		}, nil
	}
	return nil, nil
}
