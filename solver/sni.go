package solver

import (
	"context"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// sniCRUD implements crud.Actions interface.
type sniCRUD struct {
	client *kong.Client
}

func sniFromStruct(arg diff.Event) *state.SNI {
	sni, ok := arg.Obj.(*state.SNI)
	if !ok {
		panic("unexpected type, expected *state.SNI")
	}

	return sni
}

// Create creates a SNI in Kong.
// The arg should be of type diff.Event, containing the sni to be created,
// else the function will panic.
// It returns a the created *state.SNI.
func (s *sniCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	sni := sniFromStruct(event)
	createdSNI, err := s.client.SNIs.Create(ctx, &sni.SNI)
	if err != nil {
		return nil, err
	}
	return &state.SNI{SNI: *createdSNI}, nil
}

// Delete deletes a SNI in Kong.
// The arg should be of type diff.Event, containing the sni to be deleted,
// else the function will panic.
// It returns a the deleted *state.SNI.
func (s *sniCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	sni := sniFromStruct(event)
	err := s.client.SNIs.Delete(ctx, sni.ID)
	if err != nil {
		return nil, err
	}
	return sni, nil
}

// Update updates a SNI in Kong.
// The arg should be of type diff.Event, containing the sni to be updated,
// else the function will panic.
// It returns a the updated *state.SNI.
func (s *sniCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	sni := sniFromStruct(event)

	updatedSNI, err := s.client.SNIs.Create(ctx, &sni.SNI)
	if err != nil {
		return nil, err
	}
	return &state.SNI{SNI: *updatedSNI}, nil
}
