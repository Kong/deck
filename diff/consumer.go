package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteConsumers() error {
	currentConsumers, err := sc.currentState.Consumers.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching consumers from state: %w", err)
	}

	for _, consumer := range currentConsumers {
		n, err := sc.deleteConsumer(consumer)
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

func (sc *Syncer) deleteConsumer(consumer *state.Consumer) (*Event, error) {
	_, err := sc.targetState.Consumers.Get(*consumer.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "consumer",
			Obj:  consumer,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up consumer %q: %w",
			consumer.Identifier(), err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateConsumers() error {
	targetConsumers, err := sc.targetState.Consumers.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching consumers from state: %w", err)
	}

	for _, consumer := range targetConsumers {
		n, err := sc.createUpdateConsumer(consumer)
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

func (sc *Syncer) createUpdateConsumer(consumer *state.Consumer) (*Event, error) {
	consumerCopy := &state.Consumer{Consumer: *consumer.DeepCopy()}
	currentConsumer, err := sc.currentState.Consumers.Get(*consumer.ID)

	if err == state.ErrNotFound {
		// consumer not present, create it
		return &Event{
			Op:   crud.Create,
			Kind: "consumer",
			Obj:  consumerCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up consumer %q: %w",
			consumer.Identifier(), err)
	}

	// found, check if update needed
	if !currentConsumer.EqualWithOpts(consumerCopy, false, true) {
		return &Event{
			Op:     crud.Update,
			Kind:   "consumer",
			Obj:    consumerCopy,
			OldObj: currentConsumer,
		}, nil
	}
	return nil, nil
}
