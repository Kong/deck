package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
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
	if aclGroup.Consumer == nil ||
		(utils.Empty(aclGroup.Consumer.ID)) {
		return nil, errors.Errorf("acl has no associated consumer: %+v",
			*aclGroup.Group)
	}
	// first look up the consumer to get the username
	// This is needed because the targetState doesn't have ACLs indexed
	// by Consumer ID (read from file) but indexed by consumer Username
	consumer, err := sc.currentState.Consumers.Get(*aclGroup.Consumer.ID)
	if err != nil {
		return nil, errors.Wrapf(err,
			"could not find consumer '%v'", *aclGroup.Consumer.ID)
	}
	// lookup by username and Group
	_, err = sc.targetState.ACLGroups.Get(*consumer.Username,
		*aclGroup.Group)
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
		*aclGroup.Consumer.Username, *aclGroup.Group)
	if err == state.ErrNotFound {
		// aclGroup not present, create it
		// XXX fill foreign
		consumer, err := sc.currentState.Consumers.Get(*aclGroup.Consumer.Username)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not find consumer '%v' for acl %+v",
				*aclGroup.Consumer.Username, *aclGroup.Group)
		}
		aclGroup.Consumer = &consumer.Consumer
		// XXX

		aclGroup.ID = nil
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
	currentACLGroup = &state.ACLGroup{ACLGroup: *currentACLGroup.DeepCopy()}
	// found, check if update needed

	currentACLGroup.Consumer = &kong.Consumer{
		Username: currentACLGroup.Consumer.Username,
	}
	aclGroup.Consumer = &kong.Consumer{Username: aclGroup.Consumer.Username}
	if !currentACLGroup.EqualWithOpts(aclGroup, true, true, false) {
		aclGroup.ID = kong.String(*currentACLGroup.ID)

		// XXX fill foreign
		consumer, err := sc.currentState.Consumers.Get(*aclGroup.Consumer.Username)
		if err != nil {
			return nil, errors.Wrapf(err,
				"looking up consumer '%v' for acl '%v'",
				*aclGroup.Consumer.Username, *aclGroup.Group)
		}
		aclGroup.Consumer.ID = consumer.ID
		// XXX
		return &Event{
			Op:     crud.Update,
			Kind:   "acl-group",
			Obj:    aclGroup,
			OldObj: currentACLGroup,
		}, nil
	}
	return nil, nil
}
