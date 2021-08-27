package types

import (
	"context"
	"fmt"

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
	event := crud.EventFromArg(arg[0])
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
	event := crud.EventFromArg(arg[0])
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
	event := crud.EventFromArg(arg[0])
	target := targetFromStruct(event)
	// Targets in Kong cannot be updated
	err := s.client.Targets.Delete(ctx, target.Upstream.ID, target.ID)
	if err != nil {
		return nil, err
	}
	createdTarget, err := s.client.Targets.Create(ctx,
		target.Upstream.ID, &target.Target)
	if err != nil {
		return nil, err
	}
	return &state.Target{Target: *createdTarget}, nil
}

type targetDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *targetDiffer) Deletes(handler func(crud.Event) error) error {
	currentTargets, err := d.currentState.Targets.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching targets from state: %w", err)
	}

	for _, target := range currentTargets {
		n, err := d.deleteTarget(target)
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

func (d *targetDiffer) deleteTarget(target *state.Target) (*crud.Event, error) {
	_, err := d.targetState.Targets.Get(*target.Upstream.ID,
		*target.Target.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  target,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up target %q: %w",
			*target.Target.Target, err)
	}
	return nil, nil
}

func (d *targetDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetTargets, err := d.targetState.Targets.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching targets from state: %w", err)
	}

	for _, target := range targetTargets {
		n, err := d.createUpdateTarget(target)
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

func (d *targetDiffer) createUpdateTarget(target *state.Target) (*crud.Event, error) {
	target = &state.Target{Target: *target.DeepCopy()}
	currentTarget, err := d.currentState.Targets.Get(*target.Upstream.ID,
		*target.Target.ID)
	if err == state.ErrNotFound {
		// target not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  target,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up target %q: %w",
			*target.Target.Target, err)
	}
	// found, check if update needed

	if !currentTarget.EqualWithOpts(target, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    target,
			OldObj: currentTarget,
		}, nil
	}
	return nil, nil
}
