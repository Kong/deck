package solver

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
	"github.com/hbagdi/go-kong/kong"
)

// targetCRUD implements crud.Actions interface.
type targetCRUD struct {
	client *kong.Client
}

func targetFromStuct(arg diff.Event) *state.Target {
	target, ok := arg.Obj.(*state.Target)
	if !ok {
		panic("unexpected type, expected *state.Target")
	}

	return target
}

// Create creates a Target in Kong.
// The arg should be of type diff.Event, containing the target to be created,
// else the function will panic.
// It returns a the created *state.Target.
func (s *targetCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	target := targetFromStuct(event)
	createdTarget, err := s.client.Targets.Create(nil,
		target.Upstream.ID, &target.Target)
	if err != nil {
		return nil, err
	}
	return &state.Target{Target: *createdTarget}, nil
}

// Delete deletes a Target in Kong.
// The arg should be of type diff.Event, containing the target to be deleted,
// else the function will panic.
// It returns a the deleted *state.Target.
func (s *targetCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	target := targetFromStuct(event)
	err := s.client.Targets.Delete(nil, target.Upstream.ID, target.ID)
	if err != nil {
		return nil, err
	}
	return target, nil
}

// Update updates a Target in Kong.
// The arg should be of type diff.Event, containing the target to be updated,
// else the function will panic.
// It returns a the updated *state.Target.
func (s *targetCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	target := targetFromStuct(event)
	// Targets in Kong cannot be updated
	err := s.client.Targets.Delete(nil, target.Upstream.ID, target.ID)
	if err != nil {
		return nil, err
	}
	target.ID = nil
	createdTarget, err := s.client.Targets.Create(nil,
		target.Upstream.ID, &target.Target)
	if err != nil {
		return nil, err
	}
	return &state.Target{Target: *createdTarget}, nil
}
