package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteRBACRoles() error {
	currentRBACRoles, err := sc.currentState.RBACRoles.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching roles from state")
	}

	for _, role := range currentRBACRoles {
		n, err := sc.deleteRBACRole(role)
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

func (sc *Syncer) deleteRBACRole(role *state.RBACRole) (*Event, error) {
	_, err := sc.targetState.RBACRoles.Get(*role.Name)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "rbac-role",
			Obj:  role,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up rbac role '%v'",
			role.ID)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateRBACRoles() error {
	targetRBACRoles, err := sc.targetState.RBACRoles.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching rbac roles from state")
	}

	for _, role := range targetRBACRoles {
		n, err := sc.createUpdateRBACRole(role)
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

func (sc *Syncer) createUpdateRBACRole(role *state.RBACRole) (*Event, error) {
	roleCopy := &state.RBACRole{RBACRole: *role.DeepCopy()}
	currentRole, err := sc.currentState.RBACRoles.Get(*role.Name)

	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Create,
			Kind: "rbac-role",
			Obj:  roleCopy,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up rbac role %v",
			*role.ID)
	}

	// found, check if update needed
	if !currentRole.EqualWithOpts(roleCopy, false, true, false) {
		return &Event{
			Op:     crud.Update,
			Kind:   "rbac-role",
			Obj:    roleCopy,
			OldObj: currentRole,
		}, nil
	}
	return nil, nil
}
