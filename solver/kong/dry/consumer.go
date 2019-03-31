package dry

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
)

// ConsumerCRUD implements Actions interface
// from the github.com/kong/crud package for the Consumer entitiy of Kong.
type ConsumerCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func consumerFromStuct(a diff.Event) *state.Consumer {
	Consumer, ok := a.Obj.(*state.Consumer)
	if !ok {
		panic("unexpected type, expected *state.Consumer")
	}

	return Consumer
}

// Create creates a fake Consumer.
// The arg should be of type diff.Event, containing the Consumer to be created,
// else the function will panic.
// It returns a the created *state.Consumer.
func (s *ConsumerCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	Consumer := consumerFromStuct(event)

	print.CreatePrintln("creating consumer", *Consumer.Username)
	Consumer.ID = kong.String(utils.UUID())
	return Consumer, nil
}

// Delete deletes a fake Consumer.
// The arg should be of type diff.Event, containing the Consumer to be deleted,
// else the function will panic.
// It returns a the deleted *state.Consumer.
func (s *ConsumerCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	Consumer := consumerFromStuct(event)

	print.DeletePrintln("deleting consumer", *Consumer.Username)
	return Consumer, nil
}

// Update updates a fake Consumer.
// The arg should be of type diff.Event, containing the Consumer to be updated,
// else the function will panic.
// It returns a the updated *state.Consumer.
func (s *ConsumerCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	Consumer := consumerFromStuct(event)
	oldConsumerObj, ok := event.OldObj.(*state.Consumer)
	if !ok {
		panic("unexpected type, expected *state.Consumer")
	}
	oldConsumer := oldConsumerObj.DeepCopy()
	// TODO remove this hack
	oldConsumer.CreatedAt = nil
	diff, err := getDiff(oldConsumer, &Consumer.Consumer)
	if err != nil {
		return nil, err
	}
	print.UpdatePrintf("updating consumer %s\n%s", *Consumer.Username, diff)
	return Consumer, nil
}
