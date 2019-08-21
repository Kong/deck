package kong

import (
	"context"
	"encoding/json"
)

// Oauth2Service handles oauth2 credentials in Kong.
type Oauth2Service service

// Create creates an oauth2 credential in Kong
// If an ID is specified, it will be used to
// create a oauth2 credential in Kong, otherwise an ID
// is auto-generated.
func (s *Oauth2Service) Create(ctx context.Context,
	consumerUsernameOrID *string,
	oauth2Cred *Oauth2Credential) (*Oauth2Credential, error) {

	cred, err := s.client.credentials.Create(ctx, "oauth2",
		consumerUsernameOrID, oauth2Cred)
	if err != nil {
		return nil, err
	}

	var createdOauth2Cred Oauth2Credential
	err = json.Unmarshal(cred, &createdOauth2Cred)
	if err != nil {
		return nil, err
	}

	return &createdOauth2Cred, nil
}

// Get fetches an oauth2 credential from Kong.
func (s *Oauth2Service) Get(ctx context.Context,
	consumerUsernameOrID, clientIDorID *string) (*Oauth2Credential, error) {

	cred, err := s.client.credentials.Get(ctx, "oauth2",
		consumerUsernameOrID, clientIDorID)
	if err != nil {
		return nil, err
	}

	var oauth2Cred Oauth2Credential
	err = json.Unmarshal(cred, &oauth2Cred)
	if err != nil {
		return nil, err
	}

	return &oauth2Cred, nil
}

// Update updates an oauth2 credential in Kong.
func (s *Oauth2Service) Update(ctx context.Context,
	consumerUsernameOrID *string,
	oauth2Cred *Oauth2Credential) (*Oauth2Credential, error) {

	cred, err := s.client.credentials.Update(ctx, "oauth2",
		consumerUsernameOrID, oauth2Cred)
	if err != nil {
		return nil, err
	}

	var updatedHMACAuth Oauth2Credential
	err = json.Unmarshal(cred, &updatedHMACAuth)
	if err != nil {
		return nil, err
	}

	return &updatedHMACAuth, nil
}

// Delete deletes an oauth2 credential in Kong.
func (s *Oauth2Service) Delete(ctx context.Context,
	consumerUsernameOrID, clientIDorID *string) error {
	return s.client.credentials.Delete(ctx, "oauth2",
		consumerUsernameOrID, clientIDorID)
}

// List fetches a list of oauth2 credentials in Kong.
// opt can be used to control pagination.
func (s *Oauth2Service) List(ctx context.Context,
	opt *ListOpt) ([]*Oauth2Credential, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/oauth2", opt)
	if err != nil {
		return nil, nil, err
	}
	var oauth2Creds []*Oauth2Credential
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var oauth2Cred Oauth2Credential
		err = json.Unmarshal(b, &oauth2Cred)
		if err != nil {
			return nil, nil, err
		}
		oauth2Creds = append(oauth2Creds, &oauth2Cred)
	}

	return oauth2Creds, next, nil
}

// ListAll fetches all oauth2 credentials in Kong.
// This method can take a while if there
// a lot of oauth2 credentials present.
func (s *Oauth2Service) ListAll(
	ctx context.Context) ([]*Oauth2Credential, error) {
	var oauth2Creds, data []*Oauth2Credential
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		oauth2Creds = append(oauth2Creds, data...)
	}
	return oauth2Creds, nil
}

// ListForConsumer fetches a list of oauth2 credentials
// in Kong associated with a specific consumer.
// opt can be used to control pagination.
func (s *Oauth2Service) ListForConsumer(ctx context.Context,
	consumerUsernameOrID *string, opt *ListOpt) ([]*Oauth2Credential,
	*ListOpt, error) {
	data, next, err := s.client.list(ctx,
		"/consumers/"+*consumerUsernameOrID+"/oauth2", opt)
	if err != nil {
		return nil, nil, err
	}
	var oauth2Creds []*Oauth2Credential
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var oauth2Cred Oauth2Credential
		err = json.Unmarshal(b, &oauth2Cred)
		if err != nil {
			return nil, nil, err
		}
		oauth2Creds = append(oauth2Creds, &oauth2Cred)
	}

	return oauth2Creds, next, nil
}
