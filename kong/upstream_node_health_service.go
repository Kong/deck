package kong

import (
	"context"
	"encoding/json"
	"fmt"
)

// UpstreamNodeHealthService handles Upstream Node Healths in Kong.
type UpstreamNodeHealthService service

// List fetches a list of Upstream Node Healths in Kong.
// opt can be used to control pagination.
func (s *UpstreamNodeHealthService) List(
	ctx context.Context,
	upstreamNameOrID *string,
	opt *ListOpt,
) ([]*UpstreamNodeHealth, *ListOpt, error) {
	endpoint := fmt.Sprintf("/upstreams/%v/health", *upstreamNameOrID)
	data, next, err := s.client.list(ctx, endpoint, opt)
	if err != nil {
		return nil, nil, err
	}
	var upstreamNodeHealths []*UpstreamNodeHealth

	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var upstreamNodeHealth UpstreamNodeHealth
		err = json.Unmarshal(b, &upstreamNodeHealth)
		if err != nil {
			return nil, nil, err
		}
		upstreamNodeHealths = append(upstreamNodeHealths, &upstreamNodeHealth)
	}

	return upstreamNodeHealths, next, nil
}

// ListAll fetches all Upstream Node Healths in Kong.
// This method can take a while if there are
// a lot of Upstream Node Healths present.
func (s *UpstreamNodeHealthService) ListAll(
	ctx context.Context,
	upstreamNameOrID *string,
) ([]*UpstreamNodeHealth, error) {
	var upstreamNodeHealths, data []*UpstreamNodeHealth
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, upstreamNameOrID, opt)
		if err != nil {
			return nil, err
		}
		upstreamNodeHealths = append(upstreamNodeHealths, data...)
	}
	return upstreamNodeHealths, nil
}
