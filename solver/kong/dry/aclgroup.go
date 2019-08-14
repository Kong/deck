package dry

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
)

// ACLGroupCRUD implements Actions interface
// from the github.com/kong/crud package for the acl of Kong.
type ACLGroupCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func aclGroupFromStruct(arg diff.Event) *state.ACLGroup {
	aclGroup, ok := arg.Obj.(*state.ACLGroup)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return aclGroup
}

// Create creates a fake ACLGroup.
// The arg should be of type diff.Event, containing the aclGroup to be created,
// else the function will panic.
// It returns a the created *state.ACLGroup.
func (s *ACLGroupCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	aclGroup := aclGroupFromStruct(event)
	print.CreatePrintln("creating acl", *aclGroup.Group,
		"for consumer", *aclGroup.Consumer.Username)
	aclGroup.ID = kong.String(utils.UUID())
	return aclGroup, nil
}

// Delete deletes a fake Route.
// The arg should be of type diff.Event, containing the aclGroup to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *ACLGroupCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	aclGroup := aclGroupFromStruct(event)
	print.DeletePrintln("deleting acl", *aclGroup.Group,
		"for consumer", *aclGroup.Consumer.Username)
	return aclGroup, nil
}

// Update updates a fake Route.
// The arg should be of type diff.Event, containing the aclGroup to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *ACLGroupCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	aclGroup := aclGroupFromStruct(event)
	oldRoute, ok := event.OldObj.(*state.ACLGroup)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}
	// TODO remove this hack
	oldRoute.CreatedAt = nil
	oldRoute.Consumer = &kong.Consumer{Username: oldRoute.Consumer.Username}
	oldRoute.ID = nil

	aclGroup.ID = nil
	aclGroup.Consumer = &kong.Consumer{Username: aclGroup.Consumer.Username}

	diffString, err := getDiff(oldRoute.ACLGroup, aclGroup.ACLGroup)
	if err != nil {
		return nil, err
	}
	print.UpdatePrintf("updating acl %s\n%s", *aclGroup.Group, diffString)
	return aclGroup, nil
}
