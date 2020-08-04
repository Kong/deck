package solver

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/hbagdi/go-kong/kong"
)

// jwtAuthCRUD implements crud.Actions interface.
type jwtAuthCRUD struct {
	client *kong.Client
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
func (s *jwtAuthCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
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
func (s *jwtAuthCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
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
func (s *jwtAuthCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
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
