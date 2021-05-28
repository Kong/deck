package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteCertificates() error {
	currentCertificates, err := sc.currentState.Certificates.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching certificates from state: %w", err)
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
	certificate *state.Certificate) (*crud.Event, error) {
	_, err := sc.targetState.Certificates.Get(*certificate.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "certificate",
			Obj:  certificate,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up certificate %q': %w",
			certificate.Identifier(), err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateCertificates() error {
	targetCertificates, err := sc.targetState.Certificates.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching certificates from state: %w", err)
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
	certificate *state.Certificate) (*crud.Event, error) {
	certificateCopy := &state.Certificate{Certificate: *certificate.DeepCopy()}
	currentCertificate, err := sc.currentState.Certificates.Get(*certificate.ID)

	if err == state.ErrNotFound {
		// certificate not present, create it
		return &crud.Event{
			Op:   crud.Create,
			Kind: "certificate",
			Obj:  certificateCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up certificate %q: %w",
			certificate.Identifier(), err)
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
			return nil, fmt.Errorf("error looking up current certificate SNIs %q: %w",
				certificate.Identifier(), err)
		}
		sniNames := make([]*string, 0)
		for _, s := range currentSNIs {
			sniNames = append(sniNames, s.Name)
		}

		certificateCopy.SNIs = sniNames
		currentCertificate.SNIs = sniNames
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "certificate",
			Obj:    certificateCopy,
			OldObj: currentCertificate,
		}, nil
	}
	return nil, nil
}
