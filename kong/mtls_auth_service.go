package kong

import (
	"context"
	"encoding/json"
)

// MTLSAuthService handles MTLS credentials in Kong.
type MTLSAuthService service

// Create creates an MTLS credential in Kong
// If an ID is specified, it will be used to
// create a MTLS in Kong, otherwise an ID
// is auto-generated.
func (s *MTLSAuthService) Create(ctx context.Context,
	consumerUsernameOrID *string, mtlsAuth *MTLSAuth) (*MTLSAuth, error) {

	cred, err := s.client.credentials.Create(ctx, "mtls-auth",
		consumerUsernameOrID, mtlsAuth)
	if err != nil {
		return nil, err
	}

	var createdMTLS MTLSAuth
	err = json.Unmarshal(cred, &createdMTLS)
	if err != nil {
		return nil, err
	}

	return &createdMTLS, nil
}

// Get fetches an MTLS credential from Kong.
func (s *MTLSAuthService) Get(ctx context.Context,
	consumerUsernameOrID, keyOrID *string) (*MTLSAuth, error) {

	cred, err := s.client.credentials.Get(ctx, "mtls-auth",
		consumerUsernameOrID, keyOrID)
	if err != nil {
		return nil, err
	}

	var mtlsAuth MTLSAuth
	err = json.Unmarshal(cred, &mtlsAuth)
	if err != nil {
		return nil, err
	}

	return &mtlsAuth, nil
}

// Update updates an MTLS credential in Kong
func (s *MTLSAuthService) Update(ctx context.Context,
	consumerUsernameOrID *string, mtlsAuth *MTLSAuth) (*MTLSAuth, error) {

	cred, err := s.client.credentials.Update(ctx, "mtls-auth",
		consumerUsernameOrID, mtlsAuth)
	if err != nil {
		return nil, err
	}

	var updatedMTLS MTLSAuth
	err = json.Unmarshal(cred, &updatedMTLS)
	if err != nil {
		return nil, err
	}

	return &updatedMTLS, nil
}

// Delete deletes an MTLS credential in Kong
func (s *MTLSAuthService) Delete(ctx context.Context,
	consumerUsernameOrID, keyOrID *string) error {
	return s.client.credentials.Delete(ctx, "mtls-auth",
		consumerUsernameOrID, keyOrID)
}

// List fetches a list of MTLS credentials in Kong.
// opt can be used to control pagination.
func (s *MTLSAuthService) List(ctx context.Context,
	opt *ListOpt) ([]*MTLSAuth, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/mtls-auths", opt)
	if err != nil {
		return nil, nil, err
	}
	var mtlss []*MTLSAuth
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var mtlsAuth MTLSAuth
		err = json.Unmarshal(b, &mtlsAuth)
		if err != nil {
			return nil, nil, err
		}
		mtlss = append(mtlss, &mtlsAuth)
	}

	return mtlss, next, nil
}

// ListAll fetches all MTLS credentials in Kong.
// This method can take a while if there
// a lot of MTLS credentials present.
func (s *MTLSAuthService) ListAll(ctx context.Context) ([]*MTLSAuth, error) {
	var mtlss, data []*MTLSAuth
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		mtlss = append(mtlss, data...)
	}
	return mtlss, nil
}

// ListForConsumer fetches a list of mtls credentials
// in Kong associated with a specific consumer.
// opt can be used to control pagination.
func (s *MTLSAuthService) ListForConsumer(ctx context.Context,
	consumerUsernameOrID *string, opt *ListOpt) ([]*MTLSAuth, *ListOpt, error) {
	data, next, err := s.client.list(ctx,
		"/consumers/"+*consumerUsernameOrID+"/mtls-auth", opt)
	if err != nil {
		return nil, nil, err
	}
	var mtlss []*MTLSAuth
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var mtlsAuth MTLSAuth
		err = json.Unmarshal(b, &mtlsAuth)
		if err != nil {
			return nil, nil, err
		}
		mtlss = append(mtlss, &mtlsAuth)
	}

	return mtlss, next, nil
}
