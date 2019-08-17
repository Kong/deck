package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteCACertificates() error {
	currentCACertificates, err := sc.currentState.CACertificates.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching caCertificates from state")
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
	_, err := sc.targetState.CACertificates.Get(*caCert.Cert)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "ca_certificate",
			Obj:  caCert,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up caCertificate '%v'",
			*caCert.Cert)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateCACertificates() error {
	targetCACertificates, err := sc.targetState.CACertificates.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching caCertificates from state")
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
	currentCACert, err :=
		sc.currentState.CACertificates.Get(*caCert.Cert)

	if err == state.ErrNotFound {
		// caCertificate not present, create it
		caCertCopy.ID = nil
		return &Event{
			Op:   crud.Create,
			Kind: "ca_certificate",
			Obj:  caCertCopy,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up caCertificate %v",
			*caCert.Cert)
	}

	// found, check if update needed
	if !currentCACert.EqualWithOpts(caCertCopy, true, true) {
		caCertCopy.ID = kong.String(*currentCACert.ID)
		return &Event{
			Op:     crud.Update,
			Kind:   "ca_certificate",
			Obj:    caCertCopy,
			OldObj: currentCACert,
		}, nil
	}
	return nil, nil
}
