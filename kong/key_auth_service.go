package kong

import (
	"context"
	"encoding/json"
)

// KeyAuthService handles key-auth credentials in Kong.
type KeyAuthService service

// Create creates a key-auth credential in Kong
// If an ID is specified, it will be used to
// create a key-auth in Kong, otherwise an ID
// is auto-generated.
func (s *KeyAuthService) Create(ctx context.Context,
	consumerUsernameOrID *string, keyAuth *KeyAuth) (*KeyAuth, error) {

	cred, err := s.client.credentials.Create(ctx, "key-auth",
		consumerUsernameOrID, keyAuth)
	if err != nil {
		return nil, err
	}

	var createdKeyAuth KeyAuth
	err = json.Unmarshal(cred, &createdKeyAuth)
	if err != nil {
		return nil, err
	}

	return &createdKeyAuth, nil
}

// Get fetches a key-auth credential from Kong.
func (s *KeyAuthService) Get(ctx context.Context,
	consumerUsernameOrID, keyOrID *string) (*KeyAuth, error) {

	cred, err := s.client.credentials.Get(ctx, "key-auth",
		consumerUsernameOrID, keyOrID)
	if err != nil {
		return nil, err
	}

	var keyAuth KeyAuth
	err = json.Unmarshal(cred, &keyAuth)
	if err != nil {
		return nil, err
	}

	return &keyAuth, nil
}

// Update updates a key-auth credential in Kong
func (s *KeyAuthService) Update(ctx context.Context,
	consumerUsernameOrID *string, keyAuth *KeyAuth) (*KeyAuth, error) {

	cred, err := s.client.credentials.Update(ctx, "key-auth",
		consumerUsernameOrID, keyAuth)
	if err != nil {
		return nil, err
	}

	var updatedKeyAuth KeyAuth
	err = json.Unmarshal(cred, &updatedKeyAuth)
	if err != nil {
		return nil, err
	}

	return &updatedKeyAuth, nil
}

// Delete deletes a key-auth credential in Kong
func (s *KeyAuthService) Delete(ctx context.Context,
	consumerUsernameOrID, keyOrID *string) error {
	return s.client.credentials.Delete(ctx, "key-auth",
		consumerUsernameOrID, keyOrID)
}

// List fetches a list of key-auth credentials in Kong.
// opt can be used to control pagination.
func (s *KeyAuthService) List(ctx context.Context,
	opt *ListOpt) ([]*KeyAuth, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/key-auths", opt)
	if err != nil {
		return nil, nil, err
	}
	var keyAuths []*KeyAuth
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var keyAuth KeyAuth
		err = json.Unmarshal(b, &keyAuth)
		if err != nil {
			return nil, nil, err
		}
		keyAuths = append(keyAuths, &keyAuth)
	}

	return keyAuths, next, nil
}

// ListAll fetches all key-auth credentials in Kong.
// This method can take a while if there
// a lot of key-auth credentials present.
func (s *KeyAuthService) ListAll(ctx context.Context) ([]*KeyAuth, error) {
	var keyAuths, data []*KeyAuth
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		keyAuths = append(keyAuths, data...)
	}
	return keyAuths, nil
}

// ListForConsumer fetches a list of key-auth credentials
// in Kong associated with a specific consumer.
// opt can be used to control pagination.
func (s *KeyAuthService) ListForConsumer(ctx context.Context,
	consumerUsernameOrID *string, opt *ListOpt) ([]*KeyAuth, *ListOpt, error) {
	data, next, err := s.client.list(ctx,
		"/consumers/"+*consumerUsernameOrID+"/key-auth", opt)
	if err != nil {
		return nil, nil, err
	}
	var keyAuths []*KeyAuth
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var keyAuth KeyAuth
		err = json.Unmarshal(b, &keyAuth)
		if err != nil {
			return nil, nil, err
		}
		keyAuths = append(keyAuths, &keyAuth)
	}

	return keyAuths, next, nil
}
