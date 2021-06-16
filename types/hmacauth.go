package types

import (
	"context"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

// hmacAuthCRUD implements crud.Actions interface.
type hmacAuthCRUD struct {
	client *kong.Client
}

func hmacAuthFromStruct(arg crud.Event) *state.HMACAuth {
	hmacAuth, ok := arg.Obj.(*state.HMACAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return hmacAuth
}

// Create creates a Route in Kong.
// The arg should be of type crud.Event, containing the hmacAuth to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *hmacAuthCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	hmacAuth := hmacAuthFromStruct(event)
	cid := ""
	if !utils.Empty(hmacAuth.Consumer.Username) {
		cid = *hmacAuth.Consumer.Username
	}
	if !utils.Empty(hmacAuth.Consumer.ID) {
		cid = *hmacAuth.Consumer.ID
	}
	createdHMACAuth, err := s.client.HMACAuths.Create(ctx, &cid,
		&hmacAuth.HMACAuth)
	if err != nil {
		return nil, err
	}
	return &state.HMACAuth{HMACAuth: *createdHMACAuth}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type crud.Event, containing the hmacAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *hmacAuthCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	hmacAuth := hmacAuthFromStruct(event)
	cid := ""
	if !utils.Empty(hmacAuth.Consumer.Username) {
		cid = *hmacAuth.Consumer.Username
	}
	if !utils.Empty(hmacAuth.Consumer.ID) {
		cid = *hmacAuth.Consumer.ID
	}
	err := s.client.HMACAuths.Delete(ctx, &cid, hmacAuth.ID)
	if err != nil {
		return nil, err
	}
	return hmacAuth, nil
}

// Update updates a Route in Kong.
// The arg should be of type crud.Event, containing the hmacAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *hmacAuthCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	hmacAuth := hmacAuthFromStruct(event)

	cid := ""
	if !utils.Empty(hmacAuth.Consumer.Username) {
		cid = *hmacAuth.Consumer.Username
	}
	if !utils.Empty(hmacAuth.Consumer.ID) {
		cid = *hmacAuth.Consumer.ID
	}
	updatedHMACAuth, err := s.client.HMACAuths.Create(ctx, &cid, &hmacAuth.HMACAuth)
	if err != nil {
		return nil, err
	}
	return &state.HMACAuth{HMACAuth: *updatedHMACAuth}, nil
}

type hmacAuthDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *hmacAuthDiffer) Deletes(handler func(crud.Event) error) error {
	currentHMACAuths, err := d.currentState.HMACAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching hmac-auths from state: %w", err)
	}

	for _, hmacAuth := range currentHMACAuths {
		n, err := d.deleteHMACAuth(hmacAuth)
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

func (d *hmacAuthDiffer) deleteHMACAuth(hmacAuth *state.HMACAuth) (*crud.Event, error) {
	_, err := d.targetState.HMACAuths.Get(*hmacAuth.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  hmacAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up hmac-auth %q: %w",
			*hmacAuth.Username, err)
	}
	return nil, nil
}

func (d *hmacAuthDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetHMACAuths, err := d.targetState.HMACAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching hmac-auths from state: %w", err)
	}

	for _, hmacAuth := range targetHMACAuths {
		n, err := d.createUpdateHMACAuth(hmacAuth)
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

func (d *hmacAuthDiffer) createUpdateHMACAuth(hmacAuth *state.HMACAuth) (*crud.Event, error) {
	hmacAuth = &state.HMACAuth{HMACAuth: *hmacAuth.DeepCopy()}
	currentHMACAuth, err := d.currentState.HMACAuths.Get(*hmacAuth.ID)
	if err == state.ErrNotFound {
		// hmacAuth not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  hmacAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up hmac-auth %q: %w",
			*hmacAuth.Username, err)
	}
	// found, check if update needed

	if !currentHMACAuth.EqualWithOpts(hmacAuth, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    hmacAuth,
			OldObj: currentHMACAuth,
		}, nil
	}
	return nil, nil
}
