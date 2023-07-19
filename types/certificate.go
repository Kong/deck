package types

import (
	"context"
	"errors"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// certificateCRUD implements crud.Actions interface.
type certificateCRUD struct {
	client    *kong.Client
	isKonnect bool
}

func certificateFromStruct(arg crud.Event) *state.Certificate {
	certificate, ok := arg.Obj.(*state.Certificate)
	if !ok {
		panic("unexpected type, expected *state.certificate")
	}
	return certificate
}

// Create creates a Certificate in Kong.
// The arg should be of type crud.Event, containing the certificate to be created,
// else the function will panic.
// It returns a the created *state.Certificate.
func (s *certificateCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	certificate := certificateFromStruct(event)
	if s.isKonnect {
		certificate.SNIs = nil
	}
	createdCertificate, err := s.client.Certificates.Create(ctx, &certificate.Certificate)
	if err != nil {
		return nil, err
	}
	return &state.Certificate{Certificate: *createdCertificate}, nil
}

// Delete deletes a Certificate in Kong.
// The arg should be of type crud.Event, containing the certificate to be deleted,
// else the function will panic.
// It returns a the deleted *state.Certificate.
func (s *certificateCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	certificate := certificateFromStruct(event)
	err := s.client.Certificates.Delete(ctx, certificate.ID)
	if err != nil {
		return nil, err
	}
	return certificate, nil
}

// Update updates a Certificate in Kong.
// The arg should be of type crud.Event, containing the certificate to be updated,
// else the function will panic.
// It returns a the updated *state.Certificate.
func (s *certificateCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	certificate := certificateFromStruct(event)

	if s.isKonnect {
		certificate.SNIs = nil
	}
	updatedCertificate, err := s.client.Certificates.Create(ctx, &certificate.Certificate)
	if err != nil {
		return nil, err
	}
	return &state.Certificate{Certificate: *updatedCertificate}, nil
}

type certificateDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState

	isKonnect bool
}

func (d *certificateDiffer) Deletes(handler func(crud.Event) error) error {
	currentCertificates, err := d.currentState.Certificates.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching certificates from state: %w", err)
	}

	for _, certificate := range currentCertificates {
		n, err := d.deleteCertificate(certificate)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *certificateDiffer) deleteCertificate(
	certificate *state.Certificate,
) (*crud.Event, error) {
	_, err := d.targetState.Certificates.Get(*certificate.ID)
	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  certificate,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up certificate %q': %w",
			certificate.FriendlyName(), err)
	}
	return nil, nil
}

func (d *certificateDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetCertificates, err := d.targetState.Certificates.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching certificates from state: %w", err)
	}

	for _, certificate := range targetCertificates {
		n, err := d.createUpdateCertificate(certificate)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *certificateDiffer) createUpdateCertificate(
	certificate *state.Certificate,
) (*crud.Event, error) {
	certificateCopy := &state.Certificate{Certificate: *certificate.DeepCopy()}
	currentCertificate, err := d.currentState.Certificates.Get(*certificate.ID)

	if d.isKonnect {
		certificateCopy.SNIs = nil
		if currentCertificate != nil {
			currentCertificate.SNIs = nil
		}
	}

	if errors.Is(err, state.ErrNotFound) {
		// certificate not present, create it
		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  certificateCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up certificate %q: %w",
			certificate.FriendlyName(), err)
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
		if !d.isKonnect {
			currentSNIs, err := d.currentState.SNIs.GetAllByCertID(*currentCertificate.ID)
			if err != nil {
				return nil, fmt.Errorf("error looking up current certificate SNIs %q: %w",
					certificate.FriendlyName(), err)
			}
			sniNames := make([]*string, 0)
			for _, s := range currentSNIs {
				sniNames = append(sniNames, s.Name)
			}

			certificateCopy.SNIs = sniNames
			currentCertificate.SNIs = sniNames
		}
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    certificateCopy,
			OldObj: currentCertificate,
		}, nil
	}
	return nil, nil
}
