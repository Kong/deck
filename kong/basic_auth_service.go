package kong

import (
	"context"
	"encoding/json"
)

// BasicAuthService handles basic-auth credentials in Kong.
type BasicAuthService service

// Create creates a basic-auth credential in Kong
// If an ID is specified, it will be used to
// create a basic-auth in Kong, otherwise an ID
// is auto-generated.
func (s *BasicAuthService) Create(ctx context.Context,
	consumerUsernameOrID *string, basicAuth *BasicAuth) (*BasicAuth, error) {

	cred, err := s.client.credentials.Create(ctx, "basic-auth",
		consumerUsernameOrID, basicAuth)
	if err != nil {
		return nil, err
	}

	var createdBasicAuth BasicAuth
	err = json.Unmarshal(cred, &createdBasicAuth)
	if err != nil {
		return nil, err
	}

	return &createdBasicAuth, nil
}

// Get fetches a basic-auth credential from Kong.
func (s *BasicAuthService) Get(ctx context.Context,
	consumerUsernameOrID, usernameOrID *string) (*BasicAuth, error) {

	cred, err := s.client.credentials.Get(ctx, "basic-auth",
		consumerUsernameOrID, usernameOrID)
	if err != nil {
		return nil, err
	}

	var basicAuth BasicAuth
	err = json.Unmarshal(cred, &basicAuth)
	if err != nil {
		return nil, err
	}

	return &basicAuth, nil
}

// Update updates a basic-auth credential in Kong
func (s *BasicAuthService) Update(ctx context.Context,
	consumerUsernameOrID *string, basicAuth *BasicAuth) (*BasicAuth, error) {

	cred, err := s.client.credentials.Update(ctx, "basic-auth",
		consumerUsernameOrID, basicAuth)
	if err != nil {
		return nil, err
	}

	var updatedBasicAuth BasicAuth
	err = json.Unmarshal(cred, &updatedBasicAuth)
	if err != nil {
		return nil, err
	}

	return &updatedBasicAuth, nil
}

// Delete deletes a basic-auth credential in Kong
func (s *BasicAuthService) Delete(ctx context.Context,
	consumerUsernameOrID, usernameOrID *string) error {
	return s.client.credentials.Delete(ctx, "basic-auth",
		consumerUsernameOrID, usernameOrID)
}

// List fetches a list of basic-auth credentials in Kong.
// opt can be used to control pagination.
func (s *BasicAuthService) List(ctx context.Context,
	opt *ListOpt) ([]*BasicAuth, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/basic-auths", opt)
	if err != nil {
		return nil, nil, err
	}
	var basicAuths []*BasicAuth
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var basicAuth BasicAuth
		err = json.Unmarshal(b, &basicAuth)
		if err != nil {
			return nil, nil, err
		}
		basicAuths = append(basicAuths, &basicAuth)
	}

	return basicAuths, next, nil
}

// ListAll fetches all basic-auth credentials in Kong.
// This method can take a while if there
// a lot of basic-auth credentials present.
func (s *BasicAuthService) ListAll(ctx context.Context) ([]*BasicAuth, error) {
	var basicAuths, data []*BasicAuth
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		basicAuths = append(basicAuths, data...)
	}
	return basicAuths, nil
}

// ListForConsumer fetches a list of basic-auth credentials
// in Kong associated with a specific consumer.
// opt can be used to control pagination.
func (s *BasicAuthService) ListForConsumer(ctx context.Context,
	consumerUsernameOrID *string, opt *ListOpt) ([]*BasicAuth, *ListOpt, error) {
	data, next, err := s.client.list(ctx,
		"/consumers/"+*consumerUsernameOrID+"/basic-auth", opt)
	if err != nil {
		return nil, nil, err
	}
	var basicAuths []*BasicAuth
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var basicAuth BasicAuth
		err = json.Unmarshal(b, &basicAuth)
		if err != nil {
			return nil, nil, err
		}
		basicAuths = append(basicAuths, &basicAuth)
	}

	return basicAuths, next, nil
}
