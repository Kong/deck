package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteUpstreams() error {
	currentUpstreams, err := sc.currentState.Upstreams.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching upstreams from state")
	}

	for _, upstream := range currentUpstreams {
		n, err := sc.deleteUpstream(upstream)
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

func (sc *Syncer) deleteUpstream(upstream *state.Upstream) (*Event, error) {
	// lookup by name
	if utils.Empty(upstream.Name) {
		return nil, errors.New("'name' attribute for a upstream cannot be nil")
	}
	_, err := sc.targetState.Upstreams.Get(*upstream.Name)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "upstream",
			Obj:  upstream,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up upstream '%v'",
			*upstream.Name)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateUpstreams() error {
	targetUpstreams, err := sc.targetState.Upstreams.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching upstreams from state")
	}

	for _, upstream := range targetUpstreams {
		n, err := sc.createUpdateUpstream(upstream)
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

func (sc *Syncer) createUpdateUpstream(upstream *state.Upstream) (*Event,
	error) {
	upstreamCopy := &state.Upstream{Upstream: *upstream.DeepCopy()}
	currentUpstream, err := sc.currentState.Upstreams.Get(*upstream.Name)

	if err == state.ErrNotFound {
		// upstream not present, create it
		upstreamCopy.ID = nil
		return &Event{
			Op:   crud.Create,
			Kind: "upstream",
			Obj:  upstreamCopy,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up upstream %v",
			*upstream.Name)
	}

	// found, check if update needed
	if !currentUpstream.EqualWithOpts(upstreamCopy, true, true) {
		upstreamCopy.ID = kong.String(*currentUpstream.ID)
		return &Event{
			Op:     crud.Update,
			Kind:   "upstream",
			Obj:    upstreamCopy,
			OldObj: currentUpstream,
		}, nil
	}
	return nil, nil
}
