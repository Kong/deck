package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteRBACRoles() error {
	currentRBACRoles, err := sc.currentState.RBACRoles.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching rbac roles from state: %w", err)
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
		return nil, fmt.Errorf("looking up rbac role %q: %w",
			role.Identifier(), err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateRBACRoles() error {
	targetRBACRoles, err := sc.targetState.RBACRoles.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching rbac roles from state: %w", err)
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
		return nil, fmt.Errorf("error looking up rbac role %q: %w",
			role.Identifier(), err)
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
