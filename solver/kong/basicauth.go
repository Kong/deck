package kong

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

// BasicAuthCRUD implements Actions interface
// from the github.com/kong/crud package for the Route entitiy of Kong.
type BasicAuthCRUD struct {
	client *kong.Client
}

// NewBasicAuthCRUD creates a new BasicAuthCRUD. Client is required.
func NewBasicAuthCRUD(client *kong.Client) (*BasicAuthCRUD, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}
	return &BasicAuthCRUD{
		client: client,
	}, nil
}

func basicAuthFromStuct(arg diff.Event) *state.BasicAuth {
	basicAuth, ok := arg.Obj.(*state.BasicAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return basicAuth
}

// Create creates a Route in Kong.
// The arg should be of type diff.Event, containing the basicAuth to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *BasicAuthCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	basicAuth := basicAuthFromStuct(event)
	cid := ""
	if !utils.Empty(basicAuth.Consumer.Username) {
		cid = *basicAuth.Consumer.Username
	}
	if !utils.Empty(basicAuth.Consumer.ID) {
		cid = *basicAuth.Consumer.ID
	}
	createdBasicAuth, err := s.client.BasicAuths.Create(nil, &cid,
		&basicAuth.BasicAuth)
	if err != nil {
		return nil, err
	}
	return &state.BasicAuth{BasicAuth: *createdBasicAuth}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type diff.Event, containing the basicAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *BasicAuthCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	basicAuth := basicAuthFromStuct(event)
	cid := ""
	if !utils.Empty(basicAuth.Consumer.Username) {
		cid = *basicAuth.Consumer.Username
	}
	if !utils.Empty(basicAuth.Consumer.ID) {
		cid = *basicAuth.Consumer.ID
	}
	err := s.client.BasicAuths.Delete(nil, &cid, basicAuth.ID)
	if err != nil {
		return nil, err
	}
	return basicAuth, nil
}

// Update updates a Route in Kong.
// The arg should be of type diff.Event, containing the basicAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *BasicAuthCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	basicAuth := basicAuthFromStuct(event)

	cid := ""
	if !utils.Empty(basicAuth.Consumer.Username) {
		cid = *basicAuth.Consumer.Username
	}
	if !utils.Empty(basicAuth.Consumer.ID) {
		cid = *basicAuth.Consumer.ID
	}
	updatedBasicAuth, err := s.client.BasicAuths.Create(nil, &cid, &basicAuth.BasicAuth)
	if err != nil {
		return nil, err
	}
	return &state.BasicAuth{BasicAuth: *updatedBasicAuth}, nil
}
