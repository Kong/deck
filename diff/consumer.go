package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteConsumers() error {
	currentConsumers, err := sc.currentState.Consumers.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching consumers from state")
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
		return nil, errors.Wrapf(err, "looking up consumer '%v'",
			consumer.Identifier())
	}
	return nil, nil
}

func (sc *Syncer) createUpdateConsumers() error {
	targetConsumers, err := sc.targetState.Consumers.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching consumers from state")
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
		return nil, errors.Wrapf(err, "error looking up consumer %v",
			consumer.Identifier())
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
