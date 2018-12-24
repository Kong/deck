package kong

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// RouteService handles routes in Kong.
type RouteService service

// Create creates a Route in Kong
// If an ID is specified, it will be used to
// create a route in Kong, otherwise an ID
// is auto-generated.
func (s *RouteService) Create(ctx context.Context,
	route *Route) (*Route, error) {

	if route == nil {
		return nil, errors.New("cannot create a nil route")
	}

	endpoint := "/routes"
	method := "POST"
	if route.ID != nil {
		endpoint = endpoint + "/" + *route.ID
		method = "PUT"
	}
	req, err := s.client.newRequest(method, endpoint, nil, route)
	if err != nil {
		return nil, err
	}

	var createdRoute Route
	_, err = s.client.Do(ctx, req, &createdRoute)
	if err != nil {
		return nil, err
	}
	return &createdRoute, nil
}

// CreateInService creates a route associated with serviceID
func (s *RouteService) CreateInService(ctx context.Context,
	serviceID *string, route *Route) (*Route, error) {
	if isEmptyString(serviceID) {
		return nil, errors.New("serviceID cannot be nil for creating a route")
	}
	if route == nil {
		return nil, errors.New("cannot create a nil route")
	}
	r := *route
	r.Service = &Service{ID: serviceID}
	return s.Create(ctx, &r)
}

// Get fetches a Route in Kong.
func (s *RouteService) Get(ctx context.Context,
	nameOrID *string) (*Route, error) {

	if isEmptyString(nameOrID) {
		return nil, errors.New("nameOrID cannot be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/routes/%v", *nameOrID)
	req, err := s.client.newRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var route Route
	_, err = s.client.Do(ctx, req, &route)
	if err != nil {
		return nil, err
	}
	return &route, nil
}

// Update updates a Route in Kong
func (s *RouteService) Update(ctx context.Context,
	route *Route) (*Route, error) {

	if route == nil {
		return nil, errors.New("cannot update a nil route")
	}

	if isEmptyString(route.ID) {
		return nil, errors.New("ID cannot be nil for Update operation")
	}

	endpoint := fmt.Sprintf("/routes/%v", *route.ID)
	req, err := s.client.newRequest("PATCH", endpoint, nil, route)
	if err != nil {
		return nil, err
	}

	var updatedRoute Route
	_, err = s.client.Do(ctx, req, &updatedRoute)
	if err != nil {
		return nil, err
	}
	return &updatedRoute, nil
}

// Delete deletes a Route in Kong
func (s *RouteService) Delete(ctx context.Context, nameOrID *string) error {

	if isEmptyString(nameOrID) {
		return errors.New("nameOrID cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/routes/%v", *nameOrID)
	req, err := s.client.newRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}

// List fetches a list of Routes in Kong.
// opt can be used to control pagination.
func (s *RouteService) List(ctx context.Context,
	opt *ListOpt) ([]*Route, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/routes", opt)
	if err != nil {
		return nil, nil, err
	}
	var routes []*Route
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var route Route
		err = json.Unmarshal(b, &route)
		if err != nil {
			return nil, nil, err
		}
		routes = append(routes, &route)
	}

	return routes, next, nil
}

// ListAll fetches all Routes in Kong.
// This method can take a while if there
// a lot of Routes present.
func (s *RouteService) ListAll(ctx context.Context) ([]*Route, error) {
	var routes, data []*Route
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		routes = append(routes, data...)
	}
	return routes, nil
}

// ListForService fetches a list of Routes in Kong associated with a service.
// opt can be used to control pagination.
func (s *RouteService) ListForService(ctx context.Context,
	serviceNameOrID *string, opt *ListOpt) ([]*Route, *ListOpt, error) {
	data, next, err := s.client.list(ctx,
		"/services/"+*serviceNameOrID+"/routes", opt)
	if err != nil {
		return nil, nil, err
	}
	var routes []*Route
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var route Route
		err = json.Unmarshal(b, &route)
		if err != nil {
			return nil, nil, err
		}
		routes = append(routes, &route)
	}

	return routes, next, nil
}
