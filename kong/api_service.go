package kong

import (
	"context"
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
