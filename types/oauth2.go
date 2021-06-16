package types

import (
	"context"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

// oauth2CredCRUD implements crud.Actions interface.
type oauth2CredCRUD struct {
	client *kong.Client
}

func oauth2CredFromStruct(arg crud.Event) *state.Oauth2Credential {
	oauth2Cred, ok := arg.Obj.(*state.Oauth2Credential)
	if !ok {
		panic("unexpected type, expected *state.OAuth2")
	}

	return oauth2Cred
}

// Create creates a Route in Kong.
// The arg should be of type crud.Event, containing the oauth2Cred to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *oauth2CredCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	oauth2Cred := oauth2CredFromStruct(event)
	cid := ""
	if !utils.Empty(oauth2Cred.Consumer.Username) {
		cid = *oauth2Cred.Consumer.Username
	}
	if !utils.Empty(oauth2Cred.Consumer.ID) {
		cid = *oauth2Cred.Consumer.ID
	}
	createdOauth2Cred, err := s.client.Oauth2Credentials.Create(ctx, &cid,
		&oauth2Cred.Oauth2Credential)
	if err != nil {
		return nil, err
	}
	return &state.Oauth2Credential{Oauth2Credential: *createdOauth2Cred}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type crud.Event, containing the oauth2Cred to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *oauth2CredCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	oauth2Cred := oauth2CredFromStruct(event)
	cid := ""
	if !utils.Empty(oauth2Cred.Consumer.Username) {
		cid = *oauth2Cred.Consumer.Username
	}
	if !utils.Empty(oauth2Cred.Consumer.ID) {
		cid = *oauth2Cred.Consumer.ID
	}
	err := s.client.Oauth2Credentials.Delete(ctx, &cid, oauth2Cred.ID)
	if err != nil {
		return nil, err
	}
	return oauth2Cred, nil
}

// Update updates a Route in Kong.
// The arg should be of type crud.Event, containing the oauth2Cred to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *oauth2CredCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	oauth2Cred := oauth2CredFromStruct(event)

	cid := ""
	if !utils.Empty(oauth2Cred.Consumer.Username) {
		cid = *oauth2Cred.Consumer.Username
	}
	if !utils.Empty(oauth2Cred.Consumer.ID) {
		cid = *oauth2Cred.Consumer.ID
	}
	updatedOauth2Cred, err := s.client.Oauth2Credentials.Create(ctx, &cid,
		&oauth2Cred.Oauth2Credential)
	if err != nil {
		return nil, err
	}
	return &state.Oauth2Credential{Oauth2Credential: *updatedOauth2Cred}, nil
}

type oauth2CredDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *oauth2CredDiffer) Deletes(handler func(crud.Event) error) error {
	currentOauth2Creds, err := d.currentState.Oauth2Creds.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching oauth2-cred from state: %w", err)
	}

	for _, oauth2Cred := range currentOauth2Creds {
		n, err := d.deleteOauth2Cred(oauth2Cred)
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

func (d *oauth2CredDiffer) deleteOauth2Cred(oauth2Cred *state.Oauth2Credential) (
	*crud.Event, error) {
	_, err := d.targetState.Oauth2Creds.Get(*oauth2Cred.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  oauth2Cred,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up oauth2-cred %q: %w", *oauth2Cred.Name, err)
	}
	return nil, nil
}

func (d *oauth2CredDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetOauth2Creds, err := d.targetState.Oauth2Creds.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching oauth2-creds from state: %w", err)
	}

	for _, oauth2Cred := range targetOauth2Creds {
		n, err := d.createUpdateOauth2Cred(oauth2Cred)
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

func (d *oauth2CredDiffer) createUpdateOauth2Cred(oauth2Cred *state.Oauth2Credential) (*crud.Event, error) {
	oauth2Cred = &state.Oauth2Credential{Oauth2Credential: *oauth2Cred.DeepCopy()}
	currentOauth2Cred, err := d.currentState.Oauth2Creds.Get(*oauth2Cred.ID)
	if err == state.ErrNotFound {
		// oauth2Cred not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  oauth2Cred,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up oauth2-cred %q: %w",
			*oauth2Cred.Name, err)
	}
	currentOauth2Cred = &state.Oauth2Credential{Oauth2Credential: *currentOauth2Cred.DeepCopy()}
	// found, check if update needed

	if !currentOauth2Cred.EqualWithOpts(oauth2Cred, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    oauth2Cred,
			OldObj: currentOauth2Cred,
		}, nil
	}
	return nil, nil
}
