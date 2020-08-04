package solver

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

// basicAuthCRUD implements crud.Actions interface.
type basicAuthCRUD struct {
	client *kong.Client
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
func (s *basicAuthCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
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
func (s *basicAuthCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
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
func (s *basicAuthCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
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
