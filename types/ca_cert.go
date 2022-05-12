package types

import (
	"context"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// caCertificateCRUD implements crud.Actions interface.
type caCertificateCRUD struct {
	client *kong.Client
}

func caCertFromStruct(arg crud.Event) *state.CACertificate {
	caCert, ok := arg.Obj.(*state.CACertificate)
	if !ok {
		panic("unexpected type, expected *state.CACertificate")
	}
	return caCert
}

// Create creates a CACertificate in Kong.
// The arg should be of type crud.Event, containing the certificate to be created,
// else the function will panic.
// It returns a the created *state.CACertificate.
func (s *caCertificateCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	certificate := caCertFromStruct(event)
	createdCertificate, err := s.client.CACertificates.Create(ctx,
		&certificate.CACertificate)
	if err != nil {
		return nil, err
	}
	return &state.CACertificate{CACertificate: *createdCertificate}, nil
}

// Delete deletes a CACertificate in Kong.
// The arg should be of type crud.Event, containing the certificate to be deleted,
// else the function will panic.
// It returns a the deleted *state.CACertificate.
func (s *caCertificateCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	certificate := caCertFromStruct(event)
	err := s.client.CACertificates.Delete(ctx, certificate.ID)
	if err != nil {
		return nil, err
	}
	return certificate, nil
}

// Update updates a CACertificate in Kong.
// The arg should be of type crud.Event, containing the certificate to be updated,
// else the function will panic.
// It returns a the updated *state.CACertificate.
func (s *caCertificateCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	certificate := caCertFromStruct(event)
	updatedCertificate, err := s.client.CACertificates.Create(ctx,
		&certificate.CACertificate)
	if err != nil {
		return nil, err
	}
	return &state.CACertificate{CACertificate: *updatedCertificate}, nil
}

type caCertificateDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *caCertificateDiffer) Deletes(handler func(crud.Event) error) error {
	currentCACertificates, err := d.currentState.CACertificates.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching caCertificates from state: %w", err)
	}

	for _, certificate := range currentCACertificates {
		n, err := d.deleteCACertificate(certificate)
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

func (d *caCertificateDiffer) deleteCACertificate(
	caCert *state.CACertificate,
) (*crud.Event, error) {
	_, err := d.targetState.CACertificates.Get(*caCert.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  caCert,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up caCertificate %q: %w",
			caCert.FriendlyName(), err)
	}
	return nil, nil
}

func (d *caCertificateDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetCACertificates, err := d.targetState.CACertificates.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching caCertificates from state: %w", err)
	}

	for _, caCert := range targetCACertificates {
		n, err := d.createUpdateCACertificate(caCert)
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

func (d *caCertificateDiffer) createUpdateCACertificate(
	caCert *state.CACertificate,
) (*crud.Event, error) {
	caCertCopy := &state.CACertificate{CACertificate: *caCert.DeepCopy()}
	currentCACert, err := d.currentState.CACertificates.Get(*caCert.ID)

	if err == state.ErrNotFound {
		// caCertificate not present, create it
		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  caCertCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up caCertificate %q: %w",
			caCert.FriendlyName(), err)
	}

	// found, check if update needed
	if !currentCACert.EqualWithOpts(caCertCopy, false, true) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    caCertCopy,
			OldObj: currentCACert,
		}, nil
	}
	return nil, nil
}
