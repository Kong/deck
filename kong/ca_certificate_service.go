package kong

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// CACertificateService handles Certificates in Kong.
type CACertificateService service

// Create creates a CACertificate in Kong.
// If an ID is specified, it will be used to
// create a certificate in Kong, otherwise an ID
// is auto-generated.
func (s *CACertificateService) Create(ctx context.Context,
	certificate *CACertificate) (*CACertificate, error) {

	queryPath := "/ca_certificates"
	method := "POST"
	if certificate.ID != nil {
		queryPath = queryPath + "/" + *certificate.ID
		method = "PUT"
	}
	req, err := s.client.NewRequest(method, queryPath, nil, certificate)

	if err != nil {
		return nil, err
	}

	var createdCACertificate CACertificate
	_, err = s.client.Do(ctx, req, &createdCACertificate)
	if err != nil {
		return nil, err
	}
	return &createdCACertificate, nil
}

// Get fetches a CACertificate in Kong.
func (s *CACertificateService) Get(ctx context.Context,
	ID *string) (*CACertificate, error) {

	if isEmptyString(ID) {
		return nil, errors.New("ID cannot be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/ca_certificates/%v", *ID)
	req, err := s.client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var certificate CACertificate
	_, err = s.client.Do(ctx, req, &certificate)
	if err != nil {
		return nil, err
	}
	return &certificate, nil
}

// Update updates a CACertificate in Kong
func (s *CACertificateService) Update(ctx context.Context,
	certificate *CACertificate) (*CACertificate, error) {

	if isEmptyString(certificate.ID) {
		return nil, errors.New("ID cannot be nil for Update op           eration")
	}

	endpoint := fmt.Sprintf("/ca_certificates/%v", *certificate.ID)
	req, err := s.client.NewRequest("PATCH", endpoint, nil, certificate)
	if err != nil {
		return nil, err
	}

	var updatedAPI CACertificate
	_, err = s.client.Do(ctx, req, &updatedAPI)
	if err != nil {
		return nil, err
	}
	return &updatedAPI, nil
}

// Delete deletes a CACertificate in Kong
func (s *CACertificateService) Delete(ctx context.Context,
	ID *string) error {

	if isEmptyString(ID) {
		return errors.New("ID cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/ca_certificates/%v", *ID)
	req, err := s.client.NewRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}

// List fetches a list of certificate in Kong.
// opt can be used to control pagination.
func (s *CACertificateService) List(ctx context.Context,
	opt *ListOpt) ([]*CACertificate, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/ca_certificates", opt)
	if err != nil {
		return nil, nil, err
	}
	var certificates []*CACertificate
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var certificate CACertificate
		err = json.Unmarshal(b, &certificate)
		if err != nil {
			return nil, nil, err
		}
		certificates = append(certificates, &certificate)
	}

	return certificates, next, nil
}

// ListAll fetches all Certificates in Kong.
// This method can take a while if there
// a lot of Certificates present.
func (s *CACertificateService) ListAll(ctx context.Context) ([]*CACertificate,
	error) {
	var certificates, data []*CACertificate
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, data...)
	}
	return certificates, nil
}
