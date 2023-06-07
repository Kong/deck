package types

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/kong/deck/cprint"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

// basicAuthCRUD implements crud.Actions interface.
type basicAuthCRUD struct {
	client *kong.Client
}

func basicAuthFromStruct(arg crud.Event) *state.BasicAuth {
	basicAuth, ok := arg.Obj.(*state.BasicAuth)
	if !ok {
		panic("unexpected type, expected *state.BasicAuth")
	}

	return basicAuth
}

// Create creates a Route in Kong.
// The arg should be of type crud.Event, containing the basicAuth to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *basicAuthCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	basicAuth := basicAuthFromStruct(event)
	cid := ""
	if !utils.Empty(basicAuth.Consumer.Username) {
		cid = *basicAuth.Consumer.Username
	}
	if !utils.Empty(basicAuth.Consumer.ID) {
		cid = *basicAuth.Consumer.ID
	}
	createdBasicAuth, err := s.client.BasicAuths.Create(ctx, &cid,
		&basicAuth.BasicAuth)
	if err != nil {
		return nil, err
	}
	return &state.BasicAuth{BasicAuth: *createdBasicAuth}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type crud.Event, containing the basicAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *basicAuthCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	basicAuth := basicAuthFromStruct(event)
	cid := ""
	if !utils.Empty(basicAuth.Consumer.Username) {
		cid = *basicAuth.Consumer.Username
	}
	if !utils.Empty(basicAuth.Consumer.ID) {
		cid = *basicAuth.Consumer.ID
	}
	err := s.client.BasicAuths.Delete(ctx, &cid, basicAuth.ID)
	if err != nil {
		return nil, err
	}
	return basicAuth, nil
}

// Update updates a Route in Kong.
// The arg should be of type crud.Event, containing the basicAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *basicAuthCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	basicAuth := basicAuthFromStruct(event)

	cid := ""
	if !utils.Empty(basicAuth.Consumer.Username) {
		cid = *basicAuth.Consumer.Username
	}
	if !utils.Empty(basicAuth.Consumer.ID) {
		cid = *basicAuth.Consumer.ID
	}
	updatedBasicAuth, err := s.client.BasicAuths.Create(ctx, &cid, &basicAuth.BasicAuth)
	if err != nil {
		return nil, err
	}
	return &state.BasicAuth{BasicAuth: *updatedBasicAuth}, nil
}

type basicAuthDiffer struct {
	kind crud.Kind
	once sync.Once

	currentState, targetState *state.KongState
}

func (d *basicAuthDiffer) warnBasicAuth() {
	const (
		basicAuthPasswordWarning = "Warning: import/export of basic-auth" +
			"credentials using decK doesn't work due to hashing of passwords in Kong."
	)
	d.once.Do(func() {
		cprint.UpdatePrintln(basicAuthPasswordWarning)
	})
}

func (d *basicAuthDiffer) Deletes(handler func(crud.Event) error) error {
	currentBasicAuths, err := d.currentState.BasicAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching basic-auths from state: %w", err)
	}

	for _, basicAuth := range currentBasicAuths {
		n, err := d.deleteBasicAuth(basicAuth)
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

func (d *basicAuthDiffer) deleteBasicAuth(basicAuth *state.BasicAuth) (*crud.Event, error) {
	d.warnBasicAuth()
	_, err := d.targetState.BasicAuths.Get(*basicAuth.ID)
	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  basicAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up basic-auth %q: %w",
			*basicAuth.Username, err)
	}
	return nil, nil
}

func (d *basicAuthDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetBasicAuths, err := d.targetState.BasicAuths.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching basic-auths from state: %w", err)
	}

	for _, basicAuth := range targetBasicAuths {
		n, err := d.createUpdateBasicAuth(basicAuth)
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

func (d *basicAuthDiffer) createUpdateBasicAuth(basicAuth *state.BasicAuth) (*crud.Event, error) {
	d.warnBasicAuth()
	basicAuth = &state.BasicAuth{BasicAuth: *basicAuth.DeepCopy()}
	currentBasicAuth, err := d.currentState.BasicAuths.Get(*basicAuth.ID)
	if errors.Is(err, state.ErrNotFound) {
		// basicAuth not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  basicAuth,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up basic-auth %q: %w",
			*basicAuth.Username, err)
	}
	// found, check if update needed

	if !currentBasicAuth.EqualWithOpts(basicAuth, false, true, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    basicAuth,
			OldObj: currentBasicAuth,
		}, nil
	}
	return nil, nil
}
