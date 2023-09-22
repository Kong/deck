package types

import (
	"context"
	"errors"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// keySetCRUD implements crud.Actions interface.
type keySetCRUD struct {
	client *kong.Client
}

func keySetFromStruct(arg crud.Event) *state.KeySet {
	set, ok := arg.Obj.(*state.KeySet)
	if !ok {
		panic("unexpected type, expected *state.KeySet")
	}
	return set
}

// Create creates a KeySet in Kong.
// The arg should be of type crud.Event, containing the KeySet to be created,
// else the function will panic.
// It returns a the created *state.KeySet.
func (s *keySetCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	set := keySetFromStruct(event)
	createdSet, err := s.client.KeySets.Create(ctx, &set.KeySet)
	if err != nil {
		return nil, err
	}
	return &state.KeySet{KeySet: *createdSet}, nil
}

// Delete deletes a KeySet in Kong.
// The arg should be of type crud.Event, containing the KeySet to be deleted,
// else the function will panic.
// It returns a the deleted *state.KeySet.
func (s *keySetCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	set := keySetFromStruct(event)
	err := s.client.KeySets.Delete(ctx, set.ID)
	if err != nil {
		return nil, err
	}
	return set, nil
}

// Update updates a KeySet in Kong.
// The arg should be of type crud.Event, containing the KeySet to be updated,
// else the function will panic.
// It returns a the updated *state.KeySet.
func (s *keySetCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	set := keySetFromStruct(event)

	updatedSet, err := s.client.KeySets.Create(ctx, &set.KeySet)
	if err != nil {
		return nil, err
	}
	return &state.KeySet{KeySet: *updatedSet}, nil
}

type keySetDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

// Deletes generates a memdb CRUD DELETE event for KeySets
// which is then consumed by the differ and used to gate Kong client calls.
func (d *keySetDiffer) Deletes(handler func(crud.Event) error) error {
	currentSetSets, err := d.currentState.KeySets.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching key-sets from state: %w", err)
	}

	for _, set := range currentSetSets {
		n, err := d.deleteKeySet(set)
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

func (d *keySetDiffer) deleteKeySet(set *state.KeySet) (*crud.Event, error) {
	_, err := d.targetState.KeySets.Get(*set.ID)
	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "key-set",
			Obj:  set,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up key-set %q: %w", *set.Name, err)
	}
	return nil, nil
}

// CreateAndUpdates generates a memdb CRUD CREATE/UPDATE event for KeySets
// which is then consumed by the differ and used to gate Kong client calls.
func (d *keySetDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetKeySets, err := d.targetState.KeySets.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching key-sets from state: %w", err)
	}

	for _, set := range targetKeySets {
		n, err := d.createUpdateKeySet(set)
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

func (d *keySetDiffer) createUpdateKeySet(set *state.KeySet) (*crud.Event,
	error,
) {
	setCopy := &state.KeySet{KeySet: *set.DeepCopy()}
	currentSet, err := d.currentState.KeySets.Get(*set.Name)

	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Create,
			Kind: "key-set",
			Obj:  setCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up key-set %v: %w", *set.Name, err)
	}

	// found, check if update needed
	if !currentSet.EqualWithOpts(setCopy, false, true) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "key-set",
			Obj:    setCopy,
			OldObj: currentSet,
		}, nil
	}
	return nil, nil
}
