package types

import (
	"context"
	"errors"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// consumerGroupConsumerCRUD implements crud.Actions interface.
type consumerGroupConsumerCRUD struct {
	client    *kong.Client
	isKonnect bool
}

func consumerGroupConsumerFromStruct(arg crud.Event) *state.ConsumerGroupConsumer {
	consumerGroup, ok := arg.Obj.(*state.ConsumerGroupConsumer)
	if !ok {
		panic("unexpected type, expected *state.ConsumerGroupConsumer")
	}
	return consumerGroup
}

// Create creates a consumerGroupConsumer in Kong.
// The arg should be of type crud.Event, containing the consumerGroupConsumer to be created,
// else the function will panic.
// It returns the created *state.consumerGroupConsumer.
func (s *consumerGroupConsumerCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	consumer := consumerGroupConsumerFromStruct(event)

	var err error
	if s.isKonnect {
		err = konnect.CreateConsumerGroupMember(
			ctx, s.client, consumer.ConsumerGroup.ID, consumer.Consumer.ID,
		)
	} else {
		_, err = s.client.ConsumerGroupConsumers.Create(ctx, consumer.ConsumerGroup.ID, consumer.Consumer.ID)
	}
	if err != nil {
		return nil, err
	}

	return &state.ConsumerGroupConsumer{
		ConsumerGroupConsumer: kong.ConsumerGroupConsumer{
			Consumer:      consumer.Consumer,
			ConsumerGroup: consumer.ConsumerGroup,
		},
	}, nil
}

// Delete deletes a consumerGroupConsumer in Kong.
// The arg should be of type crud.Event, containing the consumerGroupConsumer to be deleted,
// else the function will panic.
// It returns the deleted *state.consumerGroupConsumer.
func (s *consumerGroupConsumerCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	consumer := consumerGroupConsumerFromStruct(event)

	var err error
	if s.isKonnect {
		err = konnect.DeleteConsumerGroupMember(ctx, s.client, consumer.ConsumerGroup.ID, consumer.Consumer.ID)
	} else {
		err = s.client.ConsumerGroupConsumers.Delete(ctx, consumer.ConsumerGroup.ID, consumer.Consumer.ID)
	}
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

// Update updates a consumerGroupConsumer in Kong.
// The arg should be of type crud.Event, containing the consumerGroupConsumer to be updated,
// else the function will panic.
// It returns the updated *state.consumerGroupConsumer.
func (s *consumerGroupConsumerCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	consumer := consumerGroupConsumerFromStruct(event)

	var err error
	// delete the old member
	if s.isKonnect {
		err = konnect.DeleteConsumerGroupMember(
			ctx, s.client, consumer.ConsumerGroup.ID, consumer.Consumer.ID,
		)
	} else {
		err = s.client.ConsumerGroupConsumers.Delete(
			ctx, consumer.ConsumerGroup.ID, consumer.Consumer.ID,
		)
	}
	if err != nil {
		return nil, err
	}

	// recreate it
	if s.isKonnect {
		err = konnect.CreateConsumerGroupMember(
			ctx, s.client, consumer.ConsumerGroup.ID, consumer.Consumer.ID,
		)
	} else {
		_, err = s.client.ConsumerGroupConsumers.Create(
			ctx, consumer.ConsumerGroup.ID, consumer.Consumer.ID,
		)
	}
	if err != nil {
		return nil, err
	}

	return &state.ConsumerGroupConsumer{
		ConsumerGroupConsumer: kong.ConsumerGroupConsumer{
			Consumer:      consumer.Consumer,
			ConsumerGroup: consumer.ConsumerGroup,
		},
	}, nil
}

type consumerGroupConsumerDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *consumerGroupConsumerDiffer) Deletes(handler func(crud.Event) error) error {
	currentConsumers, err := d.currentState.ConsumerGroupConsumers.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching consumerGroupConsumers from state: %w", err)
	}

	for _, consumer := range currentConsumers {
		n, err := d.deleteConsumerGroupConsumer(consumer)
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

func (d *consumerGroupConsumerDiffer) deleteConsumerGroupConsumer(
	consumer *state.ConsumerGroupConsumer,
) (*crud.Event, error) {
	var nameOrID string
	if consumer.Consumer.ID != nil {
		nameOrID = *consumer.Consumer.ID
	} else if consumer.Consumer.Username != nil {
		nameOrID = *consumer.Consumer.Username
	} else {
		nameOrID = *consumer.Consumer.CustomID
	}
	_, err := d.targetState.ConsumerGroupConsumers.Get(
		nameOrID, *consumer.ConsumerGroup.ID,
	)
	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "consumer-group-consumer",
			Obj:  consumer,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up consumerGroupConsumer %q: %w",
			nameOrID, err)
	}
	return nil, nil
}

func (d *consumerGroupConsumerDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetConsumers, err := d.targetState.ConsumerGroupConsumers.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching consumerGroupConsumers from state: %w", err)
	}

	for _, consumer := range targetConsumers {
		n, err := d.createUpdateConsumerGroupConsumer(consumer)
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

func (d *consumerGroupConsumerDiffer) createUpdateConsumerGroupConsumer(
	consumer *state.ConsumerGroupConsumer,
) (*crud.Event, error) {
	consumerCopy := &state.ConsumerGroupConsumer{ConsumerGroupConsumer: *consumer.DeepCopy()}
	var nameOrID string
	if consumer.Consumer.ID != nil {
		nameOrID = *consumer.Consumer.ID
	} else if consumer.Consumer.Username != nil {
		nameOrID = *consumer.Consumer.Username
	} else {
		nameOrID = *consumer.Consumer.CustomID
	}
	currentConsumer, err := d.currentState.ConsumerGroupConsumers.Get(
		nameOrID, *consumer.ConsumerGroup.ID,
	)
	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Create,
			Kind: "consumer-group-consumer",
			Obj:  consumerCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up consumerGroupConsumer %v: %w",
			nameOrID, err)
	}

	// found, check if update needed
	if !currentConsumer.EqualWithOpts(consumerCopy, false, true) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "consumer-group-consumer",
			Obj:    consumerCopy,
			OldObj: currentConsumer,
		}, nil
	}
	return nil, nil
}
