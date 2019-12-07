package solver

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
)

// aclGroupCRUD implements crud.Actions interface.
type aclGroupCRUD struct {
	client *kong.Client
}

func aclGroupFromStuct(arg diff.Event) *state.ACLGroup {
	aclGroup, ok := arg.Obj.(*state.ACLGroup)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return aclGroup
}

// Create creates a Route in Kong.
// The arg should be of type diff.Event, containing the aclGroup to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *aclGroupCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	aclGroup := aclGroupFromStuct(event)
	cid := ""
	if !utils.Empty(aclGroup.Consumer.Username) {
		cid = *aclGroup.Consumer.Username
	}
	if !utils.Empty(aclGroup.Consumer.ID) {
		cid = *aclGroup.Consumer.ID
	}
	createdACLGroup, err := s.client.ACLs.Create(nil, &cid,
		&aclGroup.ACLGroup)
	if err != nil {
		return nil, err
	}
	return &state.ACLGroup{ACLGroup: *createdACLGroup}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type diff.Event, containing the aclGroup to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *aclGroupCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	aclGroup := aclGroupFromStuct(event)
	cid := ""
	if !utils.Empty(aclGroup.Consumer.Username) {
		cid = *aclGroup.Consumer.Username
	}
	if !utils.Empty(aclGroup.Consumer.ID) {
		cid = *aclGroup.Consumer.ID
	}
	err := s.client.ACLs.Delete(nil, &cid, aclGroup.ID)
	if err != nil {
		return nil, err
	}
	return aclGroup, nil
}

// Update updates a Route in Kong.
// The arg should be of type diff.Event, containing the aclGroup to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *aclGroupCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	aclGroup := aclGroupFromStuct(event)

	cid := ""
	if !utils.Empty(aclGroup.Consumer.Username) {
		cid = *aclGroup.Consumer.Username
	}
	if !utils.Empty(aclGroup.Consumer.ID) {
		cid = *aclGroup.Consumer.ID
	}
	updatedACLGroup, err := s.client.ACLs.Create(nil, &cid, &aclGroup.ACLGroup)
	if err != nil {
		return nil, err
	}
	return &state.ACLGroup{ACLGroup: *updatedACLGroup}, nil
}
