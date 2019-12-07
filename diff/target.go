package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteTargets() error {
	currentTargets, err := sc.currentState.Targets.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching targets from state")
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

func (sc *Syncer) deleteTarget(target *state.Target) (*Event, error) {
	_, err := sc.targetState.Targets.Get(*target.Upstream.ID,
		*target.Target.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "target",
			Obj:  target,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up target '%v'",
			*target.Target.Target)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateTargets() error {
	targetTargets, err := sc.targetState.Targets.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching targets from state")
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

func (sc *Syncer) createUpdateTarget(target *state.Target) (*Event, error) {
	target = &state.Target{Target: *target.DeepCopy()}
	currentTarget, err := sc.currentState.Targets.Get(*target.Upstream.ID,
		*target.Target.ID)
	if err == state.ErrNotFound {
		// target not present, create it

		return &Event{
			Op:   crud.Create,
			Kind: "target",
			Obj:  target,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up target %v",
			*target.Target.Target)
	}
	// found, check if update needed

	if !currentTarget.EqualWithOpts(target, false, true, false) {
		return &Event{
			Op:     crud.Update,
			Kind:   "target",
			Obj:    target,
			OldObj: currentTarget,
		}, nil
	}
	return nil, nil
}
