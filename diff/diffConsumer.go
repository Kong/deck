package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
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
	// lookup by name
	if utils.Empty(consumer.Username) {
		return nil, errors.New("'name' attribute for a consumer cannot be nil")
	}
	_, err := sc.targetState.Consumers.Get(*consumer.Username)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "consumer",
			Obj:  consumer,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up consumer '%v'", *consumer.Username)
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
	currentConsumer, err := sc.currentState.Consumers.Get(*consumer.Username)

	if err == state.ErrNotFound {
		// consumer not present, create it
		consumerCopy.ID = nil
		return &Event{
			Op:   crud.Create,
			Kind: "consumer",
			Obj:  consumerCopy,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up consumer %v",
			*consumer.Username)
	}

	// found, check if update needed
	if !currentConsumer.EqualWithOpts(consumerCopy, true, true) {
		consumerCopy.ID = kong.String(*currentConsumer.ID)
		return &Event{
			Op:     crud.Update,
			Kind:   "consumer",
			Obj:    consumerCopy,
			OldObj: currentConsumer,
		}, nil
	}
	return nil, nil
}
