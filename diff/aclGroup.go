package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteACLGroups() error {
	currentACLGroups, err := sc.currentState.ACLGroups.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching acls from state: %w", err)
	}

	for _, aclGroup := range currentACLGroups {
		n, err := sc.deleteACLGroup(aclGroup)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) deleteACLGroup(aclGroup *state.ACLGroup) (*Event, error) {
	// lookup by consumerID and Group
	_, err := sc.targetState.ACLGroups.Get(*aclGroup.Consumer.ID, *aclGroup.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "acl-group",
			Obj:  aclGroup,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up acl %q: %w", *aclGroup.Group, err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateACLGroups() error {
	targetACLGroups, err := sc.targetState.ACLGroups.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching acls from state: %w", err)
	}

	for _, aclGroup := range targetACLGroups {
		n, err := sc.createUpdateACLGroup(aclGroup)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) createUpdateACLGroup(aclGroup *state.ACLGroup) (*Event, error) {
	aclGroup = &state.ACLGroup{ACLGroup: *aclGroup.DeepCopy()}
	currentACLGroup, err := sc.currentState.ACLGroups.Get(
		*aclGroup.Consumer.ID, *aclGroup.ID)
	if err == state.ErrNotFound {
		// aclGroup not present, create it

		return &Event{
			Op:   crud.Create,
			Kind: "acl-group",
			Obj:  aclGroup,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up acl %q: %w",
			*aclGroup.Group, err)
	}
	// found, check if update needed

	if !currentACLGroup.EqualWithOpts(aclGroup, false, true, false) {
		return &Event{
			Op:     crud.Update,
			Kind:   "acl-group",
			Obj:    aclGroup,
			OldObj: currentACLGroup,
		}, nil
	}
	return nil, nil
}
