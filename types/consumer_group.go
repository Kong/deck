package types

import (
	"context"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// consumerGroupCRUD implements crud.Actions interface.
type consumerGroupCRUD struct {
	client *kong.Client
}

func consumerGroupFromStruct(arg crud.Event) *state.ConsumerGroup {
	consumerGroup, ok := arg.Obj.(*state.ConsumerGroup)
	if !ok {
		panic("unexpected type, expected *state.ConsumerGroup")
	}
	return consumerGroup
}

// Create creates a consumerGroup in Kong.
// The arg should be of type crud.Event, containing the consumerGroup to be created,
// else the function will panic.
// It returns the created *state.consumerGroup.
func (s *consumerGroupCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	consumerGroup := consumerGroupFromStruct(event)
	createdConsumerGroup, err := s.client.ConsumerGroups.Create(ctx, &consumerGroup.ConsumerGroup)
	if err != nil {
		return nil, err
	}
	return &state.ConsumerGroup{ConsumerGroup: *createdConsumerGroup}, nil
}

// Delete deletes a consumerGroup in Kong.
// The arg should be of type crud.Event, containing the consumerGroup to be deleted,
// else the function will panic.
// It returns a the deleted *state.consumerGroup.
func (s *consumerGroupCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	consumerGroup := consumerGroupFromStruct(event)
	err := s.client.ConsumerGroups.Delete(ctx, consumerGroup.ConsumerGroup.ID)
	if err != nil {
		return nil, err
	}
	return consumerGroup, nil
}

// Update updates a consumerGroup in Kong.
// The arg should be of type crud.Event, containing the consumerGroup to be updated,
// else the function will panic.
// It returns a the updated *state.consumerGroup.
func (s *consumerGroupCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	consumerGroup := consumerGroupFromStruct(event)

	updatedconsumerGroup, err := s.client.ConsumerGroups.Create(ctx, &consumerGroup.ConsumerGroup)
	if err != nil {
		return nil, err
	}
	return &state.ConsumerGroup{ConsumerGroup: *updatedconsumerGroup}, nil
}

type consumerGroupDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *consumerGroupDiffer) Deletes(handler func(crud.Event) error) error {
	currentconsumerGroups, err := d.currentState.ConsumerGroups.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching consumerGroups from state: %w", err)
	}

	for _, consumerGroup := range currentconsumerGroups {
		n, err := d.deleteConsumerGroup(consumerGroup)
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

func (d *consumerGroupDiffer) deleteConsumerGroup(consumerGroup *state.ConsumerGroup) (*crud.Event, error) {
	_, err := d.targetState.ConsumerGroups.Get(*consumerGroup.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "consumer-group",
			Obj:  consumerGroup,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up consumerGroup %q: %w",
			*consumerGroup.Name, err)
	}
	return nil, nil
}

func (d *consumerGroupDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetconsumerGroups, err := d.targetState.ConsumerGroups.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching consumerGroups from state: %w", err)
	}

	for _, consumerGroup := range targetconsumerGroups {
		n, err := d.createUpdateConsumerGroup(consumerGroup)
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

func (d *consumerGroupDiffer) createUpdateConsumerGroup(consumerGroup *state.ConsumerGroup) (*crud.Event,
	error,
) {
	consumerGroupCopy := &state.ConsumerGroup{ConsumerGroup: *consumerGroup.DeepCopy()}
	currentconsumerGroup, err := d.currentState.ConsumerGroups.Get(*consumerGroup.Name)

	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Create,
			Kind: "consumer-group",
			Obj:  consumerGroupCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up consumerGroup %v: %w",
			*consumerGroup.Name, err)
	}

	// found, check if update needed
	if !currentconsumerGroup.EqualWithOpts(consumerGroupCopy, false, true) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "consumer-group",
			Obj:    consumerGroupCopy,
			OldObj: currentconsumerGroup,
		}, nil
	}
	return nil, nil
}
