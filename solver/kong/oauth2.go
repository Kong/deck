package kong

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

// Oauth2CredCRUD implements Actions interface
// from the github.com/kong/crud package for the Route entitiy of Kong.
type Oauth2CredCRUD struct {
	client *kong.Client
}

// NewOauth2CredCRUD creates a new Oauth2CredCRUD. Client is required.
func NewOauth2CredCRUD(client *kong.Client) (*Oauth2CredCRUD, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}
	return &Oauth2CredCRUD{
		client: client,
	}, nil
}

func oauth2CredFromStuct(arg diff.Event) *state.Oauth2Credential {
	oauth2Cred, ok := arg.Obj.(*state.Oauth2Credential)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return oauth2Cred
}

// Create creates a Route in Kong.
// The arg should be of type diff.Event, containing the oauth2Cred to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *Oauth2CredCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	oauth2Cred := oauth2CredFromStuct(event)
	cid := ""
	if !utils.Empty(oauth2Cred.Consumer.Username) {
		cid = *oauth2Cred.Consumer.Username
	}
	if !utils.Empty(oauth2Cred.Consumer.ID) {
		cid = *oauth2Cred.Consumer.ID
	}
	createdOauth2Cred, err := s.client.Oauth2Credentials.Create(nil, &cid,
		&oauth2Cred.Oauth2Credential)
	if err != nil {
		return nil, err
	}
	return &state.Oauth2Credential{Oauth2Credential: *createdOauth2Cred}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type diff.Event, containing the oauth2Cred to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *Oauth2CredCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	oauth2Cred := oauth2CredFromStuct(event)
	cid := ""
	if !utils.Empty(oauth2Cred.Consumer.Username) {
		cid = *oauth2Cred.Consumer.Username
	}
	if !utils.Empty(oauth2Cred.Consumer.ID) {
		cid = *oauth2Cred.Consumer.ID
	}
	err := s.client.Oauth2Credentials.Delete(nil, &cid, oauth2Cred.ID)
	if err != nil {
		return nil, err
	}
	return oauth2Cred, nil
}

// Update updates a Route in Kong.
// The arg should be of type diff.Event, containing the oauth2Cred to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *Oauth2CredCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	oauth2Cred := oauth2CredFromStuct(event)

	cid := ""
	if !utils.Empty(oauth2Cred.Consumer.Username) {
		cid = *oauth2Cred.Consumer.Username
	}
	if !utils.Empty(oauth2Cred.Consumer.ID) {
		cid = *oauth2Cred.Consumer.ID
	}
	updatedOauth2Cred, err := s.client.Oauth2Credentials.Create(nil, &cid,
		&oauth2Cred.Oauth2Credential)
	if err != nil {
		return nil, err
	}
	return &state.Oauth2Credential{Oauth2Credential: *updatedOauth2Cred}, nil
}
