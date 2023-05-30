package types

import (
	"context"
	"errors"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// sniCRUD implements crud.Actions interface.
type sniCRUD struct {
	client *kong.Client
}

func sniFromStruct(arg crud.Event) *state.SNI {
	sni, ok := arg.Obj.(*state.SNI)
	if !ok {
		panic("unexpected type, expected *state.SNI")
	}

	return sni
}

// Create creates a SNI in Kong.
// The arg should be of type crud.Event, containing the sni to be created,
// else the function will panic.
// It returns a the created *state.SNI.
func (s *sniCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	sni := sniFromStruct(event)
	createdSNI, err := s.client.SNIs.Create(ctx, &sni.SNI)
	if err != nil {
		return nil, err
	}
	return &state.SNI{SNI: *createdSNI}, nil
}

// Delete deletes a SNI in Kong.
// The arg should be of type crud.Event, containing the sni to be deleted,
// else the function will panic.
// It returns a the deleted *state.SNI.
func (s *sniCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	sni := sniFromStruct(event)
	err := s.client.SNIs.Delete(ctx, sni.ID)
	if err != nil {
		return nil, err
	}
	return sni, nil
}

// Update updates a SNI in Kong.
// The arg should be of type crud.Event, containing the sni to be updated,
// else the function will panic.
// It returns a the updated *state.SNI.
func (s *sniCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	sni := sniFromStruct(event)

	updatedSNI, err := s.client.SNIs.Create(ctx, &sni.SNI)
	if err != nil {
		return nil, err
	}
	return &state.SNI{SNI: *updatedSNI}, nil
}

type sniDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *sniDiffer) Deletes(handler func(crud.Event) error) error {
	currentSNIs, err := d.currentState.SNIs.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching snis from state: %w", err)
	}

	for _, sni := range currentSNIs {
		n, err := d.deleteSNI(sni)
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

func (d *sniDiffer) deleteSNI(sni *state.SNI) (*crud.Event, error) {
	_, err := d.targetState.SNIs.Get(*sni.ID)
	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  sni,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up sni %q: %w", *sni.Name, err)
	}
	return nil, nil
}

func (d *sniDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	sniSNIs, err := d.targetState.SNIs.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching snis from state: %w", err)
	}

	for _, sni := range sniSNIs {
		n, err := d.createUpdateSNI(sni)
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

func (d *sniDiffer) createUpdateSNI(sni *state.SNI) (*crud.Event, error) {
	sni = &state.SNI{SNI: *sni.DeepCopy()}
	currentSNI, err := d.currentState.SNIs.Get(*sni.ID)
	if errors.Is(err, state.ErrNotFound) {
		// sni not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  sni,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up sni %q: %w", *sni.Name, err)
	}
	// found, check if update needed

	if !currentSNI.EqualWithOpts(sni, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    sni,
			OldObj: currentSNI,
		}, nil
	}
	return nil, nil
}
