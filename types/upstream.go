package types

import (
	"context"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// upstreamCRUD implements crud.Actions interface.
type upstreamCRUD struct {
	client *kong.Client
}

func upstreamFromStruct(arg crud.Event) *state.Upstream {
	upstream, ok := arg.Obj.(*state.Upstream)
	if !ok {
		panic("unexpected type, expected *state.upstream")
	}
	return upstream
}

// Create creates a Upstream in Kong.
// The arg should be of type crud.Event, containing the upstream to be created,
// else the function will panic.
// It returns a the created *state.Upstream.
func (s *upstreamCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	upstream := upstreamFromStruct(event)
	createdUpstream, err := s.client.Upstreams.Create(ctx, &upstream.Upstream)
	if err != nil {
		return nil, err
	}
	return &state.Upstream{Upstream: *createdUpstream}, nil
}

// Delete deletes a Upstream in Kong.
// The arg should be of type crud.Event, containing the upstream to be deleted,
// else the function will panic.
// It returns a the deleted *state.Upstream.
func (s *upstreamCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	upstream := upstreamFromStruct(event)
	err := s.client.Upstreams.Delete(ctx, upstream.ID)
	if err != nil {
		return nil, err
	}
	return upstream, nil
}

// Update updates a Upstream in Kong.
// The arg should be of type crud.Event, containing the upstream to be updated,
// else the function will panic.
// It returns a the updated *state.Upstream.
func (s *upstreamCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	upstream := upstreamFromStruct(event)

	updatedUpstream, err := s.client.Upstreams.Create(ctx, &upstream.Upstream)
	if err != nil {
		return nil, err
	}
	return &state.Upstream{Upstream: *updatedUpstream}, nil
}

type upstreamDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *upstreamDiffer) Deletes(handler func(crud.Event) error) error {
	currentUpstreams, err := d.currentState.Upstreams.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching upstreams from state: %w", err)
	}

	for _, upstream := range currentUpstreams {
		n, err := d.deleteUpstream(upstream)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (d *upstreamDiffer) deleteUpstream(upstream *state.Upstream) (*crud.Event, error) {
	_, err := d.targetState.Upstreams.Get(*upstream.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "upstream",
			Obj:  upstream,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up upstream %q: %w",
			*upstream.Name, err)
	}
	return nil, nil
}

func (d *upstreamDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetUpstreams, err := d.targetState.Upstreams.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching upstreams from state: %w", err)
	}

	for _, upstream := range targetUpstreams {
		n, err := d.createUpdateUpstream(upstream)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *upstreamDiffer) createUpdateUpstream(upstream *state.Upstream) (*crud.Event,
	error) {
	upstreamCopy := &state.Upstream{Upstream: *upstream.DeepCopy()}
	currentUpstream, err := d.currentState.Upstreams.Get(*upstream.Name)

	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Create,
			Kind: "upstream",
			Obj:  upstreamCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up upstream %v: %w",
			*upstream.Name, err)
	}

	// found, check if update needed
	if !currentUpstream.EqualWithOpts(upstreamCopy, false, true) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "upstream",
			Obj:    upstreamCopy,
			OldObj: currentUpstream,
		}, nil
	}
	return nil, nil
}
