package kong

import (
	"context"
	"encoding/json"
)

// HMACAuthService handles hmac-auth credentials in Kong.
type HMACAuthService service

// Create creates a hmac-auth credential in Kong
// If an ID is specified, it will be used to
// create a hmac-auth in Kong, otherwise an ID
// is auto-generated.
func (s *HMACAuthService) Create(ctx context.Context,
	consumerUsernameOrID *string, hmacAuth *HMACAuth) (*HMACAuth, error) {

	cred, err := s.client.credentials.Create(ctx, "hmac-auth",
		consumerUsernameOrID, hmacAuth)
	if err != nil {
		return nil, err
	}

	var createdHMACAuth HMACAuth
	err = json.Unmarshal(cred, &createdHMACAuth)
	if err != nil {
		return nil, err
	}

	return &createdHMACAuth, nil
}

// Get fetches a hmac-auth credential from Kong.
func (s *HMACAuthService) Get(ctx context.Context,
	consumerUsernameOrID, usernameOrID *string) (*HMACAuth, error) {

	cred, err := s.client.credentials.Get(ctx, "hmac-auth",
		consumerUsernameOrID, usernameOrID)
	if err != nil {
		return nil, err
	}

	var hmacAuth HMACAuth
	err = json.Unmarshal(cred, &hmacAuth)
	if err != nil {
		return nil, err
	}

	return &hmacAuth, nil
}

// Update updates a hmac-auth credential in Kong
func (s *HMACAuthService) Update(ctx context.Context,
	consumerUsernameOrID *string, hmacAuth *HMACAuth) (*HMACAuth, error) {

	cred, err := s.client.credentials.Update(ctx, "hmac-auth",
		consumerUsernameOrID, hmacAuth)
	if err != nil {
		return nil, err
	}

	var updatedHMACAuth HMACAuth
	err = json.Unmarshal(cred, &updatedHMACAuth)
	if err != nil {
		return nil, err
	}

	return &updatedHMACAuth, nil
}

// Delete deletes a hmac-auth credential in Kong
func (s *HMACAuthService) Delete(ctx context.Context,
	consumerUsernameOrID, usernameOrID *string) error {
	return s.client.credentials.Delete(ctx, "hmac-auth",
		consumerUsernameOrID, usernameOrID)
}

// List fetches a list of hmac-auth credentials in Kong.
// opt can be used to control pagination.
func (s *HMACAuthService) List(ctx context.Context,
	opt *ListOpt) ([]*HMACAuth, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/hmac-auths", opt)
	if err != nil {
		return nil, nil, err
	}
	var hmacAuths []*HMACAuth
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var hmacAuth HMACAuth
		err = json.Unmarshal(b, &hmacAuth)
		if err != nil {
			return nil, nil, err
		}
		hmacAuths = append(hmacAuths, &hmacAuth)
	}

	return hmacAuths, next, nil
}

// ListAll fetches all hmac-auth credentials in Kong.
// This method can take a while if there
// a lot of hmac-auth credentials present.
func (s *HMACAuthService) ListAll(ctx context.Context) ([]*HMACAuth, error) {
	var hmacAuths, data []*HMACAuth
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		hmacAuths = append(hmacAuths, data...)
	}
	return hmacAuths, nil
}

// ListForConsumer fetches a list of hmac-auth credentials
// in Kong associated with a specific consumer.
// opt can be used to control pagination.
func (s *HMACAuthService) ListForConsumer(ctx context.Context,
	consumerUsernameOrID *string, opt *ListOpt) ([]*HMACAuth, *ListOpt, error) {
	data, next, err := s.client.list(ctx,
		"/consumers/"+*consumerUsernameOrID+"/hmac-auth", opt)
	if err != nil {
		return nil, nil, err
	}
	var hmacAuths []*HMACAuth
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var hmacAuth HMACAuth
		err = json.Unmarshal(b, &hmacAuth)
		if err != nil {
			return nil, nil, err
		}
		hmacAuths = append(hmacAuths, &hmacAuth)
	}

	return hmacAuths, next, nil
}
