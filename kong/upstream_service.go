package kong

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// UpstreamService handles Upstreams in Kong.
type UpstreamService service

// Create creates a Upstream in Kong.
// If an ID is specified, it will be used to
// create a upstream in Kong, otherwise an ID
// is auto-generated.
func (s *UpstreamService) Create(ctx context.Context,
	upstream *Upstream) (*Upstream, error) {

	queryPath := "/upstreams"
	method := "POST"
	if upstream.ID != nil {
		queryPath = queryPath + "/" + *upstream.ID
		method = "PUT"
	}
	req, err := s.client.newRequest(method, queryPath, nil, upstream)

	if err != nil {
		return nil, err
	}

	var createdUpstream Upstream
	_, err = s.client.Do(ctx, req, &createdUpstream)
	if err != nil {
		return nil, err
	}
	return &createdUpstream, nil
}

// Get fetches a Upstream in Kong.
func (s *UpstreamService) Get(ctx context.Context,
	upstreamNameOrID *string) (*Upstream, error) {

	if isEmptyString(upstreamNameOrID) {
		return nil, errors.New("upstreamNameOrID cannot" +
			" be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/upstreams/%v", *upstreamNameOrID)
	req, err := s.client.newRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var upstream Upstream
	_, err = s.client.Do(ctx, req, &upstream)
	if err != nil {
		return nil, err
	}
	return &upstream, nil
}

// Update updates a Upstream in Kong
func (s *UpstreamService) Update(ctx context.Context,
	upstream *Upstream) (*Upstream, error) {

	if isEmptyString(upstream.ID) {
		return nil, errors.New("ID cannot be nil for Update operation")
	}

	endpoint := fmt.Sprintf("/upstreams/%v", *upstream.ID)
	req, err := s.client.newRequest("PATCH", endpoint, nil, upstream)
	if err != nil {
		return nil, err
	}

	var updatedUpstream Upstream
	_, err = s.client.Do(ctx, req, &updatedUpstream)
	if err != nil {
		return nil, err
	}
	return &updatedUpstream, nil
}

// Delete deletes a Upstream in Kong
func (s *UpstreamService) Delete(ctx context.Context,
	upstreamNameOrID *string) error {

	if isEmptyString(upstreamNameOrID) {
		return errors.New("upstreamNameOrID cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/upstreams/%v", *upstreamNameOrID)
	req, err := s.client.newRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}

// List fetches a list of Upstreams in Kong.
// opt can be used to control pagination.
func (s *UpstreamService) List(ctx context.Context,
	opt *ListOpt) ([]*Upstream, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/upstreams", opt)
	if err != nil {
		return nil, nil, err
	}
	var upstreams []*Upstream

	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var upstream Upstream
		err = json.Unmarshal(b, &upstream)
		if err != nil {
			return nil, nil, err
		}
		upstreams = append(upstreams, &upstream)
	}

	return upstreams, next, nil
}

// ListAll fetches all Upstreams in Kong.
// This method can take a while if there
// a lot of Upstreams present.
func (s *UpstreamService) ListAll(ctx context.Context) ([]*Upstream, error) {
	var upstreams, data []*Upstream
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		upstreams = append(upstreams, data...)
	}
	return upstreams, nil
}
