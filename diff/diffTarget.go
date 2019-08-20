package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
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
	if target.Upstream == nil ||
		(utils.Empty(target.Upstream.ID)) {
		return nil, errors.Errorf("target has no associated upstream: %+v",
			target)
	}
	// lookup by Name
	_, err := sc.targetState.Targets.GetByUpstreamNameAndTarget(*target.Upstream.Name, *target.Target.Target)
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
	currentTarget, err := sc.currentState.Targets.GetByUpstreamNameAndTarget(*target.Upstream.Name, *target.Target.Target)
	if err == state.ErrNotFound {
		// target not present, create it

		// XXX fill foreign
		svc, err := sc.currentState.Upstreams.Get(*target.Upstream.Name)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not find upstream '%v' for target %+v",
				*target.Upstream.Name, *target.Target.Target)
		}
		target.Upstream = &svc.Upstream
		// XXX

		target.ID = nil
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
	currentTarget = &state.Target{Target: *currentTarget.DeepCopy()}
	// found, check if update needed

	currentTarget.Upstream = &kong.Upstream{Name: currentTarget.Upstream.Name}
	target.Upstream = &kong.Upstream{Name: target.Upstream.Name}
	if !currentTarget.EqualWithOpts(target, true, true, false) {
		target.ID = kong.String(*currentTarget.ID)

		// XXX fill foreign
		svc, err := sc.currentState.Upstreams.Get(*target.Upstream.Name)
		if err != nil {
			return nil, errors.Wrapf(err,
				"looking up upstream '%v' for target '%v'",
				*target.Upstream.Name, *target.Target.Target)
		}
		target.Upstream.ID = svc.ID
		// XXX
		return &Event{
			Op:     crud.Update,
			Kind:   "target",
			Obj:    target,
			OldObj: currentTarget,
		}, nil
	}
	return nil, nil
}
