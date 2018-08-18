package kong

import (
	"context"
	"errors"
	"fmt"
	"log"
)

// Svcservice handles services in Kong.
type Svcservice service

// Create creates an Service in Kong
// If an ID is specified, it will be used to
// create a consumer in Kong, otherwise an ID
// is auto-generated.
func (s *Svcservice) Create(ctx context.Context, service *Service) (*Service, error) {

	if service == nil {
		return nil, errors.New("cannot create a nil service")
	}

	endpoint := "/services"
	method := "POST"
	if service.ID != nil {
		endpoint = endpoint + "/" + *service.ID
		method = "PUT"
	}
	req, err := s.client.newRequest(method, endpoint, nil, service)
	if err != nil {
		return nil, err
	}

	var createdService Service
	_, err = s.client.Do(ctx, req, &createdService)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &createdService, nil
}

// Get fetches an Service in Kong.
func (s *Svcservice) Get(ctx context.Context, nameOrID *string) (*Service, error) {

	if nameOrID == nil {
		return nil, errors.New("nameOrID cannot be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/services/%v", *nameOrID)
	req, err := s.client.newRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var Service Service
	_, err = s.client.Do(ctx, req, &Service)
	if err != nil {
		return nil, err
	}
	return &Service, nil
}

// Update updates an Service in Kong
func (s *Svcservice) Update(ctx context.Context, service *Service) (*Service, error) {

	if service == nil {
		return nil, errors.New("cannot update a nil service")
	}

	if isEmptyString(service.ID) {
		return nil, errors.New("ID cannot be nil for Update operation")
	}

	endpoint := fmt.Sprintf("/services/%v", *service.ID)
	req, err := s.client.newRequest("PATCH", endpoint, nil, service)
	if err != nil {
		return nil, err
	}

	var updatedService Service
	_, err = s.client.Do(ctx, req, &updatedService)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &updatedService, nil
}

// Delete deletes an Service in Kong
func (s *Svcservice) Delete(ctx context.Context, nameOrID *string) error {

	if nameOrID == nil {
		return errors.New("nameOrID cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/services/%v", *nameOrID)
	req, err := s.client.newRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}
