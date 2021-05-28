package types

import (
	"context"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// targetCRUD implements crud.Actions interface.
type targetCRUD struct {
	client *kong.Client
}

func targetFromStruct(arg crud.Event) *state.Target {
	target, ok := arg.Obj.(*state.Target)
	if !ok {
		panic("unexpected type, expected *state.Target")
	}

	return target
}

// Create creates a Target in Kong.
// The arg should be of type crud.Event, containing the target to be created,
// else the function will panic.
// It returns a the created *state.Target.
func (s *targetCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	target := targetFromStruct(event)
	createdTarget, err := s.client.Targets.Create(ctx,
		target.Upstream.ID, &target.Target)
	if err != nil {
		return nil, err
	}
	return &state.Target{Target: *createdTarget}, nil
}

// Delete deletes a Target in Kong.
// The arg should be of type crud.Event, containing the target to be deleted,
// else the function will panic.
// It returns a the deleted *state.Target.
func (s *targetCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	target := targetFromStruct(event)
	err := s.client.Targets.Delete(ctx, target.Upstream.ID, target.ID)
	if err != nil {
		return nil, err
	}
	return target, nil
}

// Update updates a Target in Kong.
// The arg should be of type crud.Event, containing the target to be updated,
// else the function will panic.
// It returns a the updated *state.Target.
func (s *targetCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	target := targetFromStruct(event)
	// Targets in Kong cannot be updated
	err := s.client.Targets.Delete(ctx, target.Upstream.ID, target.ID)
	if err != nil {
		return nil, err
	}
	target.ID = nil
	createdTarget, err := s.client.Targets.Create(ctx,
		target.Upstream.ID, &target.Target)
	if err != nil {
		return nil, err
	}
	return &state.Target{Target: *createdTarget}, nil
}
