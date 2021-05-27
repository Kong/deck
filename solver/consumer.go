package solver

import (
	"context"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// consumerCRUD implements crud.Actions interface.
type consumerCRUD struct {
	client *kong.Client
}

func consumerFromStruct(arg diff.Event) *state.Consumer {
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
func (s *consumerCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	consumer := consumerFromStruct(event)
	createdConsumer, err := s.client.Consumers.Create(ctx, &consumer.Consumer)
	if err != nil {
		return nil, err
	}
	return &state.Consumer{Consumer: *createdConsumer}, nil
}

// Delete deletes a Consumer in Kong.
// The arg should be of type diff.Event, containing the consumer to be deleted,
// else the function will panic.
// It returns a the deleted *state.Consumer.
func (s *consumerCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	consumer := consumerFromStruct(event)
	err := s.client.Consumers.Delete(ctx, consumer.ID)
	if err != nil {
		return nil, err
	}
	return consumer, nil
}

// Update updates a Consumer in Kong.
// The arg should be of type diff.Event, containing the consumer to be updated,
// else the function will panic.
// It returns a the updated *state.Consumer.
func (s *consumerCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	consumer := consumerFromStruct(event)

	updatedConsumer, err := s.client.Consumers.Create(ctx, &consumer.Consumer)
	if err != nil {
		return nil, err
	}
	return &state.Consumer{Consumer: *updatedConsumer}, nil
}
