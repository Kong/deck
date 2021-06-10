package types

import (
	"context"

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
		panic("unexpected type, expected *state.Route")
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
