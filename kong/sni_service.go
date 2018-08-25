package kong

import (
	"context"
	"errors"
	"fmt"
)

// SNIService handles SNIs in Kong.
type SNIService service

// Create creates a SNI in Kong.
// If an ID is specified, it will be used to
// create a sni in Kong, otherwise an ID
// is auto-generated.
func (s *SNIService) Create(ctx context.Context, sni *SNI) (*SNI, error) {

	queryPath := "/snis"
	method := "POST"
	if sni.ID != nil {
		queryPath = queryPath + "/" + *sni.ID
		method = "PUT"
	}
	req, err := s.client.newRequest(method, queryPath, nil, sni)

	if err != nil {
		return nil, err
	}

	var createdSNI SNI
	_, err = s.client.Do(ctx, req, &createdSNI)
	if err != nil {
		return nil, err
	}
	return &createdSNI, nil
}

// Get fetches a SNI in Kong.
func (s *SNIService) Get(ctx context.Context, usernameOrID *string) (*SNI, error) {

	if usernameOrID == nil {
		return nil, errors.New("usernameOrID cannot be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/snis/%v", *usernameOrID)
	req, err := s.client.newRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var sni SNI
	_, err = s.client.Do(ctx, req, &sni)
	if err != nil {
		return nil, err
	}
	return &sni, nil
}

// Update updates a SNI in Kong
func (s *SNIService) Update(ctx context.Context, sni *SNI) (*SNI, error) {

	if sni.ID == nil {
		return nil, errors.New("ID cannot be nil for Update operation")
	}

	endpoint := fmt.Sprintf("/snis/%v", *sni.ID)
	req, err := s.client.newRequest("PATCH", endpoint, nil, sni)
	if err != nil {
		return nil, err
	}

	var updatedAPI SNI
	_, err = s.client.Do(ctx, req, &updatedAPI)
	if err != nil {
		return nil, err
	}
	return &updatedAPI, nil
}

// Delete deletes a SNI in Kong
func (s *SNIService) Delete(ctx context.Context, usernameOrID *string) error {

	if usernameOrID == nil {
		return errors.New("usernameOrID cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/snis/%v", *usernameOrID)
	req, err := s.client.newRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}
