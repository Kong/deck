package kong

import (
	"context"
	"errors"
	"fmt"
	"log"
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
		log.Println(err)
		return nil, err
	}
	return &createdAPI, nil
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
