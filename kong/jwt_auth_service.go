package kong

import (
	"context"
	"encoding/json"
)

// JWTAuthService handles JWT credentials in Kong.
type JWTAuthService service

// Create creates a JWT credential in Kong
// If an ID is specified, it will be used to
// create a JWT in Kong, otherwise an ID
// is auto-generated.
func (s *JWTAuthService) Create(ctx context.Context,
	consumerUsernameOrID *string, jwtAuth *JWTAuth) (*JWTAuth, error) {

	cred, err := s.client.credentials.Create(ctx, "jwt-auth",
		consumerUsernameOrID, jwtAuth)
	if err != nil {
		return nil, err
	}

	var createdJWT JWTAuth
	err = json.Unmarshal(cred, &createdJWT)
	if err != nil {
		return nil, err
	}

	return &createdJWT, nil
}

// Get fetches a JWT credential from Kong.
func (s *JWTAuthService) Get(ctx context.Context,
	consumerUsernameOrID, keyOrID *string) (*JWTAuth, error) {

	cred, err := s.client.credentials.Get(ctx, "jwt-auth",
		consumerUsernameOrID, keyOrID)
	if err != nil {
		return nil, err
	}

	var jwtAuth JWTAuth
	err = json.Unmarshal(cred, &jwtAuth)
	if err != nil {
		return nil, err
	}

	return &jwtAuth, nil
}

// Update updates a JWT credential in Kong
func (s *JWTAuthService) Update(ctx context.Context,
	consumerUsernameOrID *string, jwtAuth *JWTAuth) (*JWTAuth, error) {

	cred, err := s.client.credentials.Update(ctx, "jwt-auth",
		consumerUsernameOrID, jwtAuth)
	if err != nil {
		return nil, err
	}

	var updatedJWT JWTAuth
	err = json.Unmarshal(cred, &updatedJWT)
	if err != nil {
		return nil, err
	}

	return &updatedJWT, nil
}

// Delete deletes a JWT credential in Kong
func (s *JWTAuthService) Delete(ctx context.Context,
	consumerUsernameOrID, keyOrID *string) error {
	return s.client.credentials.Delete(ctx, "jwt-auth",
		consumerUsernameOrID, keyOrID)
}

// List fetches a list of JWT credentials in Kong.
// opt can be used to control pagination.
func (s *JWTAuthService) List(ctx context.Context,
	opt *ListOpt) ([]*JWTAuth, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/jwts", opt)
	if err != nil {
		return nil, nil, err
	}
	var jwts []*JWTAuth
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var jwtAuth JWTAuth
		err = json.Unmarshal(b, &jwtAuth)
		if err != nil {
			return nil, nil, err
		}
		jwts = append(jwts, &jwtAuth)
	}

	return jwts, next, nil
}

// ListAll fetches all JWT credentials in Kong.
// This method can take a while if there
// a lot of JWT credentials present.
func (s *JWTAuthService) ListAll(ctx context.Context) ([]*JWTAuth, error) {
	var jwts, data []*JWTAuth
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		jwts = append(jwts, data...)
	}
	return jwts, nil
}

// ListForConsumer fetches a list of jwt credentials
// in Kong associated with a specific consumer.
// opt can be used to control pagination.
func (s *JWTAuthService) ListForConsumer(ctx context.Context,
	consumerUsernameOrID *string, opt *ListOpt) ([]*JWTAuth, *ListOpt, error) {
	data, next, err := s.client.list(ctx,
		"/consumers/"+*consumerUsernameOrID+"/jwt", opt)
	if err != nil {
		return nil, nil, err
	}
	var jwts []*JWTAuth
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var jwtAuth JWTAuth
		err = json.Unmarshal(b, &jwtAuth)
		if err != nil {
			return nil, nil, err
		}
		jwts = append(jwts, &jwtAuth)
	}

	return jwts, next, nil
}
