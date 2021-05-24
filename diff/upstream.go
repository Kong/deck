package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteUpstreams() error {
	currentUpstreams, err := sc.currentState.Upstreams.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching upstreams from state: %w", err)
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
	_, err := sc.targetState.Upstreams.Get(*upstream.ID)
	if err == state.ErrNotFound {
		return &Event{
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

func (sc *Syncer) createUpdateUpstreams() error {
	targetUpstreams, err := sc.targetState.Upstreams.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching upstreams from state: %w", err)
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
		return &Event{
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
		return &Event{
			Op:     crud.Update,
			Kind:   "upstream",
			Obj:    upstreamCopy,
			OldObj: currentUpstream,
		}, nil
	}
	return nil, nil
}
