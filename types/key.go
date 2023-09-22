package types

import (
	"context"
	"errors"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// keyCRUD implements crud.Actions interface.
type keyCRUD struct {
	client *kong.Client
}

func keyFromStruct(arg crud.Event) *state.Key {
	key, ok := arg.Obj.(*state.Key)
	if !ok {
		panic("unexpected type, expected *state.Key")
	}
	return key
}

// Create creates a Key in Kong.
// The arg should be of type crud.Event, containing the Key to be created,
// else the function will panic.
// It returns a the created *state.Key.
func (s *keyCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	key := keyFromStruct(event)
	createdKey, err := s.client.Keys.Create(ctx, &key.Key)
	if err != nil {
		return nil, err
	}
	return &state.Key{Key: *createdKey}, nil
}

// Delete deletes a Key in Kong.
// The arg should be of type crud.Event, containing the Key to be deleted,
// else the function will panic.
// It returns a the deleted *state.Key.
func (s *keyCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	key := keyFromStruct(event)
	err := s.client.Keys.Delete(ctx, key.ID)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// Update updates a Key in Kong.
// The arg should be of type crud.Event, containing the Key to be updated,
// else the function will panic.
// It returns a the updated *state.Key.
func (s *keyCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	key := keyFromStruct(event)

	updatedKey, err := s.client.Keys.Create(ctx, &key.Key)
	if err != nil {
		return nil, err
	}
	return &state.Key{Key: *updatedKey}, nil
}

type keyDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

// Deletes generates a memdb CRUD DELETE event for Keys
// which is then consumed by the differ and used to gate Kong client calls.
func (d *keyDiffer) Deletes(handler func(crud.Event) error) error {
	currentKeys, err := d.currentState.Keys.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching keys from state: %w", err)
	}

	for _, key := range currentKeys {
		n, err := d.deleteKey(key)
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

func (d *keyDiffer) deleteKey(key *state.Key) (*crud.Event, error) {
	_, err := d.targetState.Keys.Get(*key.ID)
	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "key",
			Obj:  key,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up key %q: %w", *key.Name, err)
	}
	return nil, nil
}

// CreateAndUpdates generates a memdb CRUD CREATE/UPDATE event for Keys
// which is then consumed by the differ and used to gate Kong client calls.
func (d *keyDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetKeys, err := d.targetState.Keys.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching keys from state: %w", err)
	}

	for _, key := range targetKeys {
		n, err := d.createUpdateKey(key)
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

func (d *keyDiffer) createUpdateKey(key *state.Key) (*crud.Event,
	error,
) {
	keyCopy := &state.Key{Key: *key.DeepCopy()}
	currentKey, err := d.currentState.Keys.Get(*key.Name)

	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Create,
			Kind: "key",
			Obj:  keyCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up Key %v: %w", *key.Name, err)
	}

	// found, check if update needed
	if !currentKey.EqualWithOpts(keyCopy, false, true) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "key",
			Obj:    keyCopy,
			OldObj: currentKey,
		}, nil
	}
	return nil, nil
}
