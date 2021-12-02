package types

import (
	"context"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// developerCRUD implements crud.Actions interface.
type developerCRUD struct {
	client *kong.Client
}

func developerFromStruct(arg crud.Event) *state.Developer {
	developer, ok := arg.Obj.(*state.Developer)
	if !ok {
		panic("unexpected type, expected *state.developer")
	}
	return developer
}

// Create creates a Developer in Kong.
// The arg should be of type crud.Event, containing the developer to be created,
// else the function will panic.
// It returns a the created *state.Developer.
func (s *developerCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	developer := developerFromStruct(event)
	createdDeveloper, err := s.client.Developers.Create(ctx, &developer.Developer)
	if err != nil {
		return nil, err
	}
	return &state.Developer{Developer: *createdDeveloper}, nil
}

// Delete deletes a Developer in Kong.
// The arg should be of type crud.Event, containing the developer to be deleted,
// else the function will panic.
// It returns a the deleted *state.Developer.
func (s *developerCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	developer := developerFromStruct(event)

	err := s.client.Developers.Delete(ctx, developer.ID)
	if err != nil {
		return nil, err
	}
	return developer, nil
}

// Update updates a Developer in Kong.
// The arg should be of type crud.Event, containing the developer to be updated,
// else the function will panic.
// It returns a the updated *state.Developer.
func (s *developerCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	developer := developerFromStruct(event)

	updatedDeveloper, err := s.client.Developers.Create(ctx, &developer.Developer)
	if err != nil {
		return nil, err
	}
	return &state.Developer{Developer: *updatedDeveloper}, nil
}

type developerDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *developerDiffer) Deletes(handler func(crud.Event) error) error {
	currentDevelopers, err := d.currentState.Developers.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching developers from state: %w", err)
	}

	for _, developer := range currentDevelopers {
		n, err := d.deleteDeveloper(developer)
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

func (d *developerDiffer) deleteDeveloper(developer *state.Developer) (*crud.Event, error) {
	_, err := d.targetState.Developers.Get(*developer.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  developer,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up developer %q: %w",
			developer.Identifier(), err)
	}
	return nil, nil
}

func (d *developerDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetDevelopers, err := d.targetState.Developers.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching developers from state: %w", err)
	}

	for _, developer := range targetDevelopers {
		n, err := d.createUpdateDeveloper(developer)
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

func (d *developerDiffer) createUpdateDeveloper(developer *state.Developer) (*crud.Event, error) {
	developerCopy := &state.Developer{Developer: *developer.DeepCopy()}
	currentDeveloper, err := d.currentState.Developers.Get(*developer.ID)

	if err == state.ErrNotFound {
		// developer not present, create it
		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  developerCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up developer %q: %w",
			developer.Identifier(), err)
	}

	// found, check if update needed
	if !currentDeveloper.EqualWithOpts(developerCopy, false, true, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    developerCopy,
			OldObj: currentDeveloper,
		}, nil
	}
	return nil, nil
}
