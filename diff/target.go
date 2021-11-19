package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteTargets() error {
	currentTargets, err := sc.currentState.Targets.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching targets from state")
	}
	fmt.Printf("DDB deleteTargets() found %v targets in current state", len(currentTargets))

	for _, target := range currentTargets {
		fmt.Printf("DDB diff.deleteTargets processing target: id %v, upstream %v, target %v\n", *target.ID, *target.Upstream.ID, *target.Target.Target)
		n, err := sc.deleteTarget(target)
		if err != nil {
			return err
		}
		if n != nil {
			fmt.Printf("DDB diff.deleteTargets queueing delete event: id %v, upstream %v, target %v\n", *n.Obj.(*state.Target).ID, *n.Obj.(*state.Target).Upstream.ID, *n.Obj.(*state.Target).Target.Target)
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) deleteTarget(target *state.Target) (*Event, error) {
	found, err := sc.targetState.Targets.Get(*target.Upstream.ID,
		*target.Target.ID)
	if found != nil {
		fmt.Printf("DDB diff.deleteTarget found target in target state: id %v, upstream %v, target %v\n", *found.ID, *found.Upstream.ID, *found.Target.Target)
	} else {
		fmt.Printf("DDB diff.deleteTarget did not find target in target state: id %v, upstream %v, target %v\n", *target.ID, *target.Upstream.ID, *target.Target.Target)
	}
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
