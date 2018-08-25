package kong

import (
	"context"
	"errors"
	"fmt"
)

// APIService handles APIs in Kong.
type APIService service

// Create creates an API in Kong
func (s *APIService) Create(ctx context.Context, api *API) (*API, error) {

	req, err := s.client.newRequest("POST", "/apis", nil, api)
	if err != nil {
		return nil, err
	}

	var createdAPI API
	_, err = s.client.Do(ctx, req, &createdAPI)
	if err != nil {
		return nil, err
	}
	return &createdAPI, nil
}

// Get fetches an API in Kong.
func (s *APIService) Get(ctx context.Context, nameOrID *string) (*API, error) {

	if nameOrID == nil {
		return nil, errors.New("nameOrID cannot be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/apis/%v", *nameOrID)
	req, err := s.client.newRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var api API
	_, err = s.client.Do(ctx, req, &api)
	if err != nil {
		return nil, err
	}
	return &api, nil
}

// Update updates an API in Kong
func (s *APIService) Update(ctx context.Context, api *API) (*API, error) {

	if api.ID == nil {
		return nil, errors.New("ID cannot be nil for Update operation")
	}

	endpoint := fmt.Sprintf("/apis/%v", *api.ID)
	req, err := s.client.newRequest("PATCH", endpoint, nil, api)
	if err != nil {
		return nil, err
	}

	var updatedAPI API
	_, err = s.client.Do(ctx, req, &updatedAPI)
	if err != nil {
		return nil, err
	}
	return &updatedAPI, nil
}

// Delete deletes an API in Kong
func (s *APIService) Delete(ctx context.Context, nameOrID *string) error {

	if nameOrID == nil {
		return errors.New("nameOrID cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/apis/%v", *nameOrID)
	req, err := s.client.newRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}
