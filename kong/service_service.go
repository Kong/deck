package kong

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Svcservice handles services in Kong.
type Svcservice service

// Create creates an Service in Kong
// If an ID is specified, it will be used to
// create a service in Kong, otherwise an ID
// is auto-generated.
func (s *Svcservice) Create(ctx context.Context,
	service *Service) (*Service, error) {

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
		return nil, err
	}
	return &createdService, nil
}

// Get fetches an Service in Kong.
func (s *Svcservice) Get(ctx context.Context,
	nameOrID *string) (*Service, error) {

	if isEmptyString(nameOrID) {
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

// GetForRoute fetches a Service associated with routeID in Kong.
func (s *Svcservice) GetForRoute(ctx context.Context,
	routeID *string) (*Service, error) {

	if isEmptyString(routeID) {
		return nil, errors.New("routeID cannot be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/routes/%v/service", *routeID)
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
func (s *Svcservice) Update(ctx context.Context,
	service *Service) (*Service, error) {

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
		return nil, err
	}
	return &updatedService, nil
}

// Delete deletes an Service in Kong
func (s *Svcservice) Delete(ctx context.Context, nameOrID *string) error {

	if isEmptyString(nameOrID) {
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

// List fetches a list of Services in Kong.
// opt can be used to control pagination.
func (s *Svcservice) List(ctx context.Context,
	opt *ListOpt) ([]*Service, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/services", opt)
	if err != nil {
		return nil, nil, err
	}
	var services []*Service
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var service Service
		err = json.Unmarshal(b, &service)
		if err != nil {
			return nil, nil, err
		}
		services = append(services, &service)
	}

	return services, next, nil
}

// ListAll fetches all Services in Kong.
// This method can take a while if there
// a lot of Services present.
func (s *Svcservice) ListAll(ctx context.Context) ([]*Service, error) {
	var services, data []*Service
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		services = append(services, data...)
	}
	return services, nil
}
