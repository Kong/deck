package types

import (
	"context"
	"errors"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// consumerCRUD implements crud.Actions interface.
type consumerCRUD struct {
	client *kong.Client
}

func consumerFromStruct(arg crud.Event) *state.Consumer {
	consumer, ok := arg.Obj.(*state.Consumer)
	if !ok {
		panic("unexpected type, expected *state.consumer")
	}
	return consumer
}

// Create creates a Consumer in Kong.
// The arg should be of type crud.Event, containing the consumer to be created,
// else the function will panic.
// It returns a the created *state.Consumer.
func (s *consumerCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	consumer := consumerFromStruct(event)
	createdConsumer, err := s.client.Consumers.Create(ctx, &consumer.Consumer)
	if err != nil {
		return nil, err
	}
	return &state.Consumer{Consumer: *createdConsumer}, nil
}

// Delete deletes a Consumer in Kong.
// The arg should be of type crud.Event, containing the consumer to be deleted,
// else the function will panic.
// It returns a the deleted *state.Consumer.
func (s *consumerCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	consumer := consumerFromStruct(event)
	err := s.client.Consumers.Delete(ctx, consumer.ID)
	if err != nil {
		return nil, err
	}
	return consumer, nil
}

// Update updates a Consumer in Kong.
// The arg should be of type crud.Event, containing the consumer to be updated,
// else the function will panic.
// It returns a the updated *state.Consumer.
func (s *consumerCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	consumer := consumerFromStruct(event)

	updatedConsumer, err := s.client.Consumers.Create(ctx, &consumer.Consumer)
	if err != nil {
		return nil, err
	}
	return &state.Consumer{Consumer: *updatedConsumer}, nil
}

type consumerDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *consumerDiffer) Deletes(handler func(crud.Event) error) error {
	currentConsumers, err := d.currentState.Consumers.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching consumers from state: %w", err)
	}

	for _, consumer := range currentConsumers {
		n, err := d.deleteConsumer(consumer)
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

func (d *consumerDiffer) deleteConsumer(consumer *state.Consumer) (*crud.Event, error) {
	_, err := d.targetState.Consumers.GetByIDOrUsername(*consumer.ID)
	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  consumer,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up consumer %q: %w",
			consumer.FriendlyName(), err)
	}
	return nil, nil
}

func (d *consumerDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetConsumers, err := d.targetState.Consumers.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching consumers from state: %w", err)
	}

	for _, consumer := range targetConsumers {
		n, err := d.createUpdateConsumer(consumer)
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

func (d *consumerDiffer) createUpdateConsumer(consumer *state.Consumer) (*crud.Event, error) {
	consumerCopy := &state.Consumer{Consumer: *consumer.DeepCopy()}
	currentConsumer, err := d.currentState.Consumers.GetByIDOrUsername(*consumer.ID)

	if errors.Is(err, state.ErrNotFound) {
		// consumer not present, create it
		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  consumerCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up consumer %q: %w",
			consumer.FriendlyName(), err)
	}

	// found, check if update needed
	if !currentConsumer.EqualWithOpts(consumerCopy, false, true) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    consumerCopy,
			OldObj: currentConsumer,
		}, nil
	}
	return nil, nil
}

func (d *consumerDiffer) DuplicatesDeletes() ([]crud.Event, error) {
	targetConsumers, err := d.targetState.Consumers.GetAll()
	if err != nil {
		return nil, fmt.Errorf("error fetching consumers from state: %w", err)
	}

	var events []crud.Event
	for _, targetConsumer := range targetConsumers {
		event, err := d.deleteDuplicateConsumer(targetConsumer)
		if err != nil {
			return nil, err
		}
		if event != nil {
			events = append(events, *event)
		}
	}

	return events, nil
}

func (d *consumerDiffer) deleteDuplicateConsumer(targetConsumer *state.Consumer) (*crud.Event, error) {
	currentConsumer, err := d.currentState.Consumers.GetByIDOrUsername(*targetConsumer.Username)
	if errors.Is(err, state.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up consumer %q: %w",
			*targetConsumer.Username, err)
	}

	if *currentConsumer.ID != *targetConsumer.ID {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "consumer",
			Obj:  currentConsumer,
		}, nil
	}

	return nil, nil
}
