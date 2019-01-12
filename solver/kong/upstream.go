package kong

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

// UpstreamCRUD implements Actions interface
// from the github.com/kong/crud package for the Upstream entitiy of Kong.
type UpstreamCRUD struct {
	client *kong.Client
}

// NewUpstreamCRUD creates a new UpstreamCRUD. Client is required.
func NewUpstreamCRUD(client *kong.Client) (*UpstreamCRUD, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}
	return &UpstreamCRUD{
		client: client,
	}, nil
}

func upstreamFromStuct(arg diff.Event) *state.Upstream {
	upstream, ok := arg.Obj.(*state.Upstream)
	if !ok {
		panic("unexpected type, expected *state.upstream")
	}
	return upstream
}

// Create creates a Upstream in Kong.
// The arg should be of type diff.Event, containing the upstream to be created,
// else the function will panic.
// It returns a the created *state.Upstream.
func (s *UpstreamCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	upstream := upstreamFromStuct(event)
	createdUpstream, err := s.client.Upstreams.Create(nil, &upstream.Upstream)
	if err != nil {
		return nil, err
	}
	return &state.Upstream{Upstream: *createdUpstream}, nil
}

// Delete deletes a Upstream in Kong.
// The arg should be of type diff.Event, containing the upstream to be deleted,
// else the function will panic.
// It returns a the deleted *state.Upstream.
func (s *UpstreamCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	upstream := upstreamFromStuct(event)
	err := s.client.Upstreams.Delete(nil, upstream.ID)
	if err != nil {
		return nil, err
	}
	return upstream, nil
}

// Update updates a Upstream in Kong.
// The arg should be of type diff.Event, containing the upstream to be updated,
// else the function will panic.
// It returns a the updated *state.Upstream.
func (s *UpstreamCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	upstream := upstreamFromStuct(event)

	updatedUpstream, err := s.client.Upstreams.Create(nil, &upstream.Upstream)
	if err != nil {
		return nil, err
	}
	return &state.Upstream{Upstream: *updatedUpstream}, nil
}
