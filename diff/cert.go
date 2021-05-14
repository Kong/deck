package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteCertificates() error {
	currentCertificates, err := sc.currentState.Certificates.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching certificates from state")
	}

	for _, certificate := range currentCertificates {
		n, err := sc.deleteCertificate(certificate)
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

func (sc *Syncer) deleteCertificate(
	certificate *state.Certificate) (*Event, error) {
	_, err := sc.targetState.Certificates.Get(*certificate.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "certificate",
			Obj:  certificate,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up certificate '%v'",
			certificate.Identifier())
	}
	return nil, nil
}

func (sc *Syncer) createUpdateCertificates() error {
	targetCertificates, err := sc.targetState.Certificates.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching certificates from state")
	}

	for _, certificate := range targetCertificates {
		n, err := sc.createUpdateCertificate(certificate)
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

func (sc *Syncer) createUpdateCertificate(
	certificate *state.Certificate) (*Event, error) {
	certificateCopy := &state.Certificate{Certificate: *certificate.DeepCopy()}
	currentCertificate, err := sc.currentState.Certificates.Get(*certificate.ID)

	if err == state.ErrNotFound {
		// certificate not present, create it
		return &Event{
			Op:   crud.Create,
			Kind: "certificate",
			Obj:  certificateCopy,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up certificate %v",
			certificate.Identifier())
	}

	// found, check if update needed
	if !currentCertificate.EqualWithOpts(certificateCopy, false, true) {
		// Certificate and SNI objects have a special relationship. A PUT request
		// (which we use for updates) with a certificate that contains no SNI
		// children will in fact delete any existing SNI objects associated with
		// that certificate, rather than leaving them as-is.

		// To work around this issues, we set SNIs on certificates here using the
		// current certificate's SNI list. If there are changes to the SNIs,
		// subsequent actions on the SNI objects will handle those.
		currentSNIs, err := sc.currentState.SNIs.GetAllByCertID(*currentCertificate.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "error looking up current certificate SNIs %v",
				certificate.Identifier())
		}
		sniNames := make([]*string, 0)
		for _, s := range currentSNIs {
			sniNames = append(sniNames, s.Name)
		}

		certificateCopy.SNIs = sniNames
		currentCertificate.SNIs = sniNames
		return &Event{
			Op:     crud.Update,
			Kind:   "certificate",
			Obj:    certificateCopy,
			OldObj: currentCertificate,
		}, nil
	}
	return nil, nil
}
