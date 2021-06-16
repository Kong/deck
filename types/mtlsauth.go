package types

import (
	"context"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

// mtlsAuthCRUD implements crud.Actions interface.
type mtlsAuthCRUD struct {
	client *kong.Client
}

func mtlsAuthFromStruct(arg crud.Event) *state.MTLSAuth {
	mtlsAuth, ok := arg.Obj.(*state.MTLSAuth)
	if !ok {
		panic("unexpected type, expected *state.MTLSAuth")
	}

	return mtlsAuth
}

// Create creates an mtls-auth credential in Kong.
// The arg should be of type crud.Event, containing the mtlsAuth to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *mtlsAuthCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	mtlsAuth := mtlsAuthFromStruct(event)
	createdMTLSAuth, err := s.client.MTLSAuths.Create(ctx, mtlsAuth.Consumer.ID,
		&mtlsAuth.MTLSAuth)
	if err != nil {
		return nil, err
	}
	return &state.MTLSAuth{MTLSAuth: *createdMTLSAuth}, nil
}

// Delete deletes an mtls-auth credential in Kong.
// The arg should be of type crud.Event, containing the mtlsAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *mtlsAuthCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	mtlsAuth := mtlsAuthFromStruct(event)
	cid := ""
	if !utils.Empty(mtlsAuth.Consumer.Username) {
		cid = *mtlsAuth.Consumer.Username
	}
	if !utils.Empty(mtlsAuth.Consumer.ID) {
		cid = *mtlsAuth.Consumer.ID
	}
	err := s.client.MTLSAuths.Delete(ctx, &cid, mtlsAuth.ID)
	if err != nil {
		return nil, err
	}
	return mtlsAuth, nil
}

// Update updates an mtls-auth credential in Kong.
// The arg should be of type crud.Event, containing the mtlsAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *mtlsAuthCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	mtlsAuth := mtlsAuthFromStruct(event)

	updatedMTLSAuth, err := s.client.MTLSAuths.Create(ctx, mtlsAuth.Consumer.ID,
		&mtlsAuth.MTLSAuth)
	if err != nil {
		return nil, err
	}
	return &state.MTLSAuth{MTLSAuth: *updatedMTLSAuth}, nil
}

type mtlsAuthDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *mtlsAuthDiffer) Deletes(handler func(crud.Event) error) error {
	currentMTLSAuths, err := d.currentState.MTLSAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching mtls-auths from state: %w", err)
	}

	for _, mtlsAuth := range currentMTLSAuths {
		n, err := d.deleteMTLSAuth(mtlsAuth)
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

func (d *mtlsAuthDiffer) deleteMTLSAuth(mtlsAuth *state.MTLSAuth) (*crud.Event, error) {
	_, err := d.targetState.MTLSAuths.Get(*mtlsAuth.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  mtlsAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up mtls-auth %q: %w", *mtlsAuth.ID, err)
	}
	return nil, nil
}

func (d *mtlsAuthDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetMTLSAuths, err := d.targetState.MTLSAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching mtls-auths from state: %w", err)
	}

	for _, mtlsAuth := range targetMTLSAuths {
		n, err := d.createUpdateMTLSAuth(mtlsAuth)
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

func (d *mtlsAuthDiffer) createUpdateMTLSAuth(mtlsAuth *state.MTLSAuth) (*crud.Event, error) {
	mtlsAuth = &state.MTLSAuth{MTLSAuth: *mtlsAuth.DeepCopy()}
	currentMTLSAuth, err := d.currentState.MTLSAuths.Get(*mtlsAuth.ID)
	if err == state.ErrNotFound {
		// mtlsAuth not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  mtlsAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up mtls-auth %q: %w",
			*mtlsAuth.ID, err)
	}
	// found, check if update needed

	if !currentMTLSAuth.EqualWithOpts(mtlsAuth, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    mtlsAuth,
			OldObj: currentMTLSAuth,
		}, nil
	}
	return nil, nil
}
