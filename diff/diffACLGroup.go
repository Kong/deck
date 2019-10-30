package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteACLGroups() error {
	currentACLGroups, err := sc.currentState.ACLGroups.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching acls from state")
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
		return nil, errors.Wrapf(err, "looking up acl '%v'", *aclGroup.Group)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateACLGroups() error {
	targetACLGroups, err := sc.targetState.ACLGroups.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching acls from state")
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
		return nil, errors.Wrapf(err, "error looking up acl %v",
			*aclGroup.Group)
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
