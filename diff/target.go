package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteTargets() error {
	currentTargets, err := sc.currentState.Targets.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching targets from state: %w", err)
	}

	for _, target := range currentTargets {
		n, err := sc.deleteTarget(target)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) deleteTarget(target *state.Target) (*crud.Event, error) {
	_, err := sc.targetState.Targets.Get(*target.Upstream.ID,
		*target.Target.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "target",
			Obj:  target,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up target %q: %w",
			*target.Target.Target, err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateTargets() error {
	targetTargets, err := sc.targetState.Targets.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching targets from state: %w", err)
	}

	for _, target := range targetTargets {
		n, err := sc.createUpdateTarget(target)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) createUpdateTarget(target *state.Target) (*crud.Event, error) {
	target = &state.Target{Target: *target.DeepCopy()}
	currentTarget, err := sc.currentState.Targets.Get(*target.Upstream.ID,
		*target.Target.ID)
	if err == state.ErrNotFound {
		// target not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: "target",
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
			Kind:   "target",
			Obj:    target,
			OldObj: currentTarget,
		}, nil
	}
	return nil, nil
}
