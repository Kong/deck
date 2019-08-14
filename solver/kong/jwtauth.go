package kong

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

// JWTAuthCRUD implements Actions interface
// from the github.com/kong/crud package for the Route entitiy of Kong.
type JWTAuthCRUD struct {
	client *kong.Client
}

// NewJWTAuthCRUD creates a new JWTAuthCRUD. Client is required.
func NewJWTAuthCRUD(client *kong.Client) (*JWTAuthCRUD, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}
	return &JWTAuthCRUD{
		client: client,
	}, nil
}

func jwtAuthFromStuct(arg diff.Event) *state.JWTAuth {
	jwtAuth, ok := arg.Obj.(*state.JWTAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return jwtAuth
}

// Create creates a Route in Kong.
// The arg should be of type diff.Event, containing the jwtAuth to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *JWTAuthCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	jwtAuth := jwtAuthFromStuct(event)
	cid := ""
	if !utils.Empty(jwtAuth.Consumer.Username) {
		cid = *jwtAuth.Consumer.Username
	}
	if !utils.Empty(jwtAuth.Consumer.ID) {
		cid = *jwtAuth.Consumer.ID
	}
	createdJWTAuth, err := s.client.JWTAuths.Create(nil, &cid,
		&jwtAuth.JWTAuth)
	if err != nil {
		return nil, err
	}
	return &state.JWTAuth{JWTAuth: *createdJWTAuth}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type diff.Event, containing the jwtAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *JWTAuthCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	jwtAuth := jwtAuthFromStuct(event)
	cid := ""
	if !utils.Empty(jwtAuth.Consumer.Username) {
		cid = *jwtAuth.Consumer.Username
	}
	if !utils.Empty(jwtAuth.Consumer.ID) {
		cid = *jwtAuth.Consumer.ID
	}
	err := s.client.JWTAuths.Delete(nil, &cid, jwtAuth.ID)
	if err != nil {
		return nil, err
	}
	return jwtAuth, nil
}

// Update updates a Route in Kong.
// The arg should be of type diff.Event, containing the jwtAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *JWTAuthCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	jwtAuth := jwtAuthFromStuct(event)

	cid := ""
	if !utils.Empty(jwtAuth.Consumer.Username) {
		cid = *jwtAuth.Consumer.Username
	}
	if !utils.Empty(jwtAuth.Consumer.ID) {
		cid = *jwtAuth.Consumer.ID
	}
	updatedJWTAuth, err := s.client.JWTAuths.Create(nil, &cid, &jwtAuth.JWTAuth)
	if err != nil {
		return nil, err
	}
	return &state.JWTAuth{JWTAuth: *updatedJWTAuth}, nil
}
