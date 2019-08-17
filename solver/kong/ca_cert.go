package kong

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

// CACertificateCRUD implements Actions interface
// from the github.com/kong/crud package for the Certificate entitiy of Kong.
type CACertificateCRUD struct {
	client *kong.Client
}

// NewCACertificateCRUD creates a new CACertificateCRUD. Client is required.
func NewCACertificateCRUD(client *kong.Client) (*CACertificateCRUD, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}
	return &CACertificateCRUD{
		client: client,
	}, nil
}

func caCertFromStuct(arg diff.Event) *state.CACertificate {
	caCert, ok := arg.Obj.(*state.CACertificate)
	if !ok {
		panic("unexpected type, expected *state.CACertificate")
	}
	return caCert
}

// Create creates a CACertificate in Kong.
// The arg should be of type diff.Event, containing the certificate to be created,
// else the function will panic.
// It returns a the created *state.CACertificate.
func (s *CACertificateCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	certificate := caCertFromStuct(event)
	createdCertificate, err := s.client.CACertificates.Create(nil,
		&certificate.CACertificate)
	if err != nil {
		return nil, err
	}
	return &state.CACertificate{CACertificate: *createdCertificate}, nil
}

// Delete deletes a CACertificate in Kong.
// The arg should be of type diff.Event, containing the certificate to be deleted,
// else the function will panic.
// It returns a the deleted *state.CACertificate.
func (s *CACertificateCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	certificate := caCertFromStuct(event)
	err := s.client.CACertificates.Delete(nil, certificate.ID)
	if err != nil {
		return nil, err
	}
	return certificate, nil
}

// Update updates a CACertificate in Kong.
// The arg should be of type diff.Event, containing the certificate to be updated,
// else the function will panic.
// It returns a the updated *state.CACertificate.
func (s *CACertificateCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	certificate := caCertFromStuct(event)
	updatedCertificate, err := s.client.CACertificates.Create(nil,
		&certificate.CACertificate)
	if err != nil {
		return nil, err
	}
	return &state.CACertificate{CACertificate: *updatedCertificate}, nil
}
