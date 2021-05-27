package solver

import (
	"context"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// upstreamCRUD implements crud.Actions interface.
type upstreamCRUD struct {
	client *kong.Client
}

func upstreamFromStruct(arg diff.Event) *state.Upstream {
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
func (s *upstreamCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	upstream := upstreamFromStruct(event)
	createdUpstream, err := s.client.Upstreams.Create(ctx, &upstream.Upstream)
	if err != nil {
		return nil, err
	}
	return &state.Upstream{Upstream: *createdUpstream}, nil
}

// Delete deletes a Upstream in Kong.
// The arg should be of type diff.Event, containing the upstream to be deleted,
// else the function will panic.
// It returns a the deleted *state.Upstream.
func (s *upstreamCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	upstream := upstreamFromStruct(event)
	err := s.client.Upstreams.Delete(ctx, upstream.ID)
	if err != nil {
		return nil, err
	}
	return upstream, nil
}

// Update updates a Upstream in Kong.
// The arg should be of type diff.Event, containing the upstream to be updated,
// else the function will panic.
// It returns a the updated *state.Upstream.
func (s *upstreamCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	upstream := upstreamFromStruct(event)

	updatedUpstream, err := s.client.Upstreams.Create(ctx, &upstream.Upstream)
	if err != nil {
		return nil, err
	}
	return &state.Upstream{Upstream: *updatedUpstream}, nil
}
