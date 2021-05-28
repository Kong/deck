package types

import (
	"context"

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
	event := eventFromArg(arg[0])
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
	event := eventFromArg(arg[0])
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
	event := eventFromArg(arg[0])
	certificate := caCertFromStruct(event)
	updatedCertificate, err := s.client.CACertificates.Create(ctx,
		&certificate.CACertificate)
	if err != nil {
		return nil, err
	}
	return &state.CACertificate{CACertificate: *updatedCertificate}, nil
}
