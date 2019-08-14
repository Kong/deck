package kong

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

// HMACAuthCRUD implements Actions interface
// from the github.com/kong/crud package for the Route entitiy of Kong.
type HMACAuthCRUD struct {
	client *kong.Client
}

// NewHMACAuthCRUD creates a new HMACAuthCRUD. Client is required.
func NewHMACAuthCRUD(client *kong.Client) (*HMACAuthCRUD, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}
	return &HMACAuthCRUD{
		client: client,
	}, nil
}

func hmacAuthFromStuct(arg diff.Event) *state.HMACAuth {
	hmacAuth, ok := arg.Obj.(*state.HMACAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return hmacAuth
}

// Create creates a Route in Kong.
// The arg should be of type diff.Event, containing the hmacAuth to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *HMACAuthCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	hmacAuth := hmacAuthFromStuct(event)
	cid := ""
	if !utils.Empty(hmacAuth.Consumer.Username) {
		cid = *hmacAuth.Consumer.Username
	}
	if !utils.Empty(hmacAuth.Consumer.ID) {
		cid = *hmacAuth.Consumer.ID
	}
	createdHMACAuth, err := s.client.HMACAuths.Create(nil, &cid,
		&hmacAuth.HMACAuth)
	if err != nil {
		return nil, err
	}
	return &state.HMACAuth{HMACAuth: *createdHMACAuth}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type diff.Event, containing the hmacAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *HMACAuthCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	hmacAuth := hmacAuthFromStuct(event)
	cid := ""
	if !utils.Empty(hmacAuth.Consumer.Username) {
		cid = *hmacAuth.Consumer.Username
	}
	if !utils.Empty(hmacAuth.Consumer.ID) {
		cid = *hmacAuth.Consumer.ID
	}
	err := s.client.HMACAuths.Delete(nil, &cid, hmacAuth.ID)
	if err != nil {
		return nil, err
	}
	return hmacAuth, nil
}

// Update updates a Route in Kong.
// The arg should be of type diff.Event, containing the hmacAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *HMACAuthCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	hmacAuth := hmacAuthFromStuct(event)

	cid := ""
	if !utils.Empty(hmacAuth.Consumer.Username) {
		cid = *hmacAuth.Consumer.Username
	}
	if !utils.Empty(hmacAuth.Consumer.ID) {
		cid = *hmacAuth.Consumer.ID
	}
	updatedHMACAuth, err := s.client.HMACAuths.Create(nil, &cid, &hmacAuth.HMACAuth)
	if err != nil {
		return nil, err
	}
	return &state.HMACAuth{HMACAuth: *updatedHMACAuth}, nil
}
