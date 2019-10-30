package dry

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
)

// TargetCRUD implements Actions interface
// from the github.com/kong/crud package for the Target entitiy of Kong.
type TargetCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func targetFromStuct(arg diff.Event) *state.Target {
	target, ok := arg.Obj.(*state.Target)
	if !ok {
		panic("unexpected type, expected *state.Target")
	}

	return target
}

// Create creates a fake Target.
// The arg should be of type diff.Event, containing the target to be created,
// else the function will panic.
// It returns a the created *state.Target.
func (s *TargetCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	target := targetFromStuct(event)
	print.CreatePrintln("creating target", *target.Target.Target,
		"on upstream", *target.Upstream.ID)
	return target, nil
}

// Delete deletes a fake Target.
// The arg should be of type diff.Event, containing the target to be deleted,
// else the function will panic.
// It returns a the deleted *state.Target.
func (s *TargetCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	target := targetFromStuct(event)
	print.DeletePrintln("deleting target", *target.Target.Target,
		"from upstream", *target.Upstream.ID)
	return target, nil
}

// Update updates a fake Target.
// The arg should be of type diff.Event, containing the target to be updated,
// else the function will panic.
// It returns a the updated *state.Target.
func (s *TargetCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	target := targetFromStuct(event)
	oldTarget, ok := event.OldObj.(*state.Target)
	if !ok {
		panic("unexpected type, expected *state.Target")
	}
	print.DeletePrintln("deleting target", *oldTarget.Target.Target,
		"from upstream", *oldTarget.Upstream.ID)
	print.CreatePrintln("creating target", *target.Target.Target,
		"on upstream", *target.Upstream.ID)
	return target, nil
}
