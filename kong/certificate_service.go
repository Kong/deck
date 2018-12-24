package kong

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// CertificateService handles Certificates in Kong.
type CertificateService service

// Create creates a Certificate in Kong.
// If an ID is specified, it will be used to
// create a certificate in Kong, otherwise an ID
// is auto-generated.
func (s *CertificateService) Create(ctx context.Context,
	certificate *Certificate) (*Certificate, error) {

	queryPath := "/certificates"
	method := "POST"
	if certificate.ID != nil {
		queryPath = queryPath + "/" + *certificate.ID
		method = "PUT"
	}
	req, err := s.client.newRequest(method, queryPath, nil, certificate)

	if err != nil {
		return nil, err
	}

	var createdCertificate Certificate
	_, err = s.client.Do(ctx, req, &createdCertificate)
	if err != nil {
		return nil, err
	}
	return &createdCertificate, nil
}

// Get fetches a Certificate in Kong.
func (s *CertificateService) Get(ctx context.Context,
	usernameOrID *string) (*Certificate, error) {

	if isEmptyString(usernameOrID) {
		return nil, errors.New("usernameOrID cannot be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/certificates/%v", *usernameOrID)
	req, err := s.client.newRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var certificate Certificate
	_, err = s.client.Do(ctx, req, &certificate)
	if err != nil {
		return nil, err
	}
	return &certificate, nil
}

// Update updates a Certificate in Kong
func (s *CertificateService) Update(ctx context.Context,
	certificate *Certificate) (*Certificate, error) {

	if isEmptyString(certificate.ID) {
		return nil, errors.New("ID cannot be nil for Update op           eration")
	}

	endpoint := fmt.Sprintf("/certificates/%v", *certificate.ID)
	req, err := s.client.newRequest("PATCH", endpoint, nil, certificate)
	if err != nil {
		return nil, err
	}

	var updatedAPI Certificate
	_, err = s.client.Do(ctx, req, &updatedAPI)
	if err != nil {
		return nil, err
	}
	return &updatedAPI, nil
}

// Delete deletes a Certificate in Kong
func (s *CertificateService) Delete(ctx context.Context,
	usernameOrID *string) error {

	if isEmptyString(usernameOrID) {
		return errors.New("usernameOrID cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/certificates/%v", *usernameOrID)
	req, err := s.client.newRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}

// List fetches a list of certificate in Kong.
// opt can be used to control pagination.
func (s *CertificateService) List(ctx context.Context,
	opt *ListOpt) ([]*Certificate, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/certificates", opt)
	if err != nil {
		return nil, nil, err
	}
	var certificates []*Certificate
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var certificate Certificate
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
func (s *CertificateService) ListAll(ctx context.Context) ([]*Certificate,
	error) {
	var certificates, data []*Certificate
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
