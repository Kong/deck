package types

import (
	"context"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// certificateCRUD implements crud.Actions interface.
type certificateCRUD struct {
	client *kong.Client
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
	event := eventFromArg(arg[0])
	certificate := certificateFromStruct(event)
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
	event := eventFromArg(arg[0])
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
	event := eventFromArg(arg[0])
	certificate := certificateFromStruct(event)

	updatedCertificate, err := s.client.Certificates.Create(ctx, &certificate.Certificate)
	if err != nil {
		return nil, err
	}
	return &state.Certificate{Certificate: *updatedCertificate}, nil
}
