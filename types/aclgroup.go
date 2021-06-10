package types

import (
	"context"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

// aclGroupCRUD implements crud.Actions interface.
type aclGroupCRUD struct {
	client *kong.Client
}

func aclGroupFromStruct(arg crud.Event) *state.ACLGroup {
	aclGroup, ok := arg.Obj.(*state.ACLGroup)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return aclGroup
}

// Create creates a Route in Kong.
// The arg should be of type crud.Event, containing the aclGroup to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *aclGroupCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	aclGroup := aclGroupFromStruct(event)
	cid := ""
	if !utils.Empty(aclGroup.Consumer.Username) {
		cid = *aclGroup.Consumer.Username
	}
	if !utils.Empty(aclGroup.Consumer.ID) {
		cid = *aclGroup.Consumer.ID
	}
	createdACLGroup, err := s.client.ACLs.Create(ctx, &cid,
		&aclGroup.ACLGroup)
	if err != nil {
		return nil, err
	}
	return &state.ACLGroup{ACLGroup: *createdACLGroup}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type crud.Event, containing the aclGroup to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *aclGroupCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	aclGroup := aclGroupFromStruct(event)
	cid := ""
	if !utils.Empty(aclGroup.Consumer.Username) {
		cid = *aclGroup.Consumer.Username
	}
	if !utils.Empty(aclGroup.Consumer.ID) {
		cid = *aclGroup.Consumer.ID
	}
	err := s.client.ACLs.Delete(ctx, &cid, aclGroup.ID)
	if err != nil {
		return nil, err
	}
	return aclGroup, nil
}

// Update updates a Route in Kong.
// The arg should be of type crud.Event, containing the aclGroup to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *aclGroupCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	aclGroup := aclGroupFromStruct(event)

	cid := ""
	if !utils.Empty(aclGroup.Consumer.Username) {
		cid = *aclGroup.Consumer.Username
	}
	if !utils.Empty(aclGroup.Consumer.ID) {
		cid = *aclGroup.Consumer.ID
	}
	updatedACLGroup, err := s.client.ACLs.Create(ctx, &cid, &aclGroup.ACLGroup)
	if err != nil {
		return nil, err
	}
	return &state.ACLGroup{ACLGroup: *updatedACLGroup}, nil
}
