package kong

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

// ConsumerCRUD implements Actions interface
// from the github.com/kong/crud package for the Consumer entitiy of Kong.
type ConsumerCRUD struct {
	client *kong.Client
}

// NewConsumerCRUD creates a new ConsumerCRUD. Client is required.
func NewConsumerCRUD(client *kong.Client) (*ConsumerCRUD, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}
	return &ConsumerCRUD{
		client: client,
	}, nil
}

func consumerFromStuct(arg diff.Event) *state.Consumer {
	consumer, ok := arg.Obj.(*state.Consumer)
	if !ok {
		panic("unexpected type, expected *state.consumer")
	}
	return consumer
}

// Create creates a Consumer in Kong.
// The arg should be of type diff.Event, containing the consumer to be created,
// else the function will panic.
// It returns a the created *state.Consumer.
func (s *ConsumerCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	consumer := consumerFromStuct(event)
	createdConsumer, err := s.client.Consumers.Create(nil, &consumer.Consumer)
	if err != nil {
		return nil, err
	}
	return &state.Consumer{Consumer: *createdConsumer}, nil
}

// Delete deletes a Consumer in Kong.
// The arg should be of type diff.Event, containing the consumer to be deleted,
// else the function will panic.
// It returns a the deleted *state.Consumer.
func (s *ConsumerCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	consumer := consumerFromStuct(event)
	err := s.client.Consumers.Delete(nil, consumer.ID)
	if err != nil {
		return nil, err
	}
	return consumer, nil
}

// Update updates a Consumer in Kong.
// The arg should be of type diff.Event, containing the consumer to be updated,
// else the function will panic.
// It returns a the updated *state.Consumer.
func (s *ConsumerCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	consumer := consumerFromStuct(event)

	updatedConsumer, err := s.client.Consumers.Create(nil, &consumer.Consumer)
	if err != nil {
		return nil, err
	}
	return &state.Consumer{Consumer: *updatedConsumer}, nil
}
