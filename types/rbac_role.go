package types

import (
	"context"
	"errors"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// rbacRoleCRUD implements crud.Actions interface.
type rbacRoleCRUD struct {
	client *kong.Client
}

func rbacRoleFromStruct(arg crud.Event) *state.RBACRole {
	role, ok := arg.Obj.(*state.RBACRole)
	if !ok {
		panic("unexpected type, expected *state.RBACRole")
	}

	return role
}

// Create creates a RBACRole in Kong.
// The arg should be of type crud.Event, containing the role to be created,
// else the function will panic.
// It returns a the created *state.RBACRole.
func (s *rbacRoleCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	role := rbacRoleFromStruct(event)
	createdRBACRole, err := s.client.RBACRoles.Create(ctx, &role.RBACRole)
	if err != nil {
		return nil, err
	}
	return &state.RBACRole{RBACRole: *createdRBACRole}, nil
}

// Delete deletes a RBACRole in Kong.
// The arg should be of type crud.Event, containing the role to be deleted,
// else the function will panic.
// It returns a the deleted *state.RBACRole.
func (s *rbacRoleCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	role := rbacRoleFromStruct(event)
	err := s.client.RBACRoles.Delete(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	return role, nil
}

// Update updates a RBACRole in Kong.
// The arg should be of type crud.Event, containing the role to be updated,
// else the function will panic.
// It returns a the updated *state.RBACRole.
func (s *rbacRoleCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	role := rbacRoleFromStruct(event)

	updatedRBACRole, err := s.client.RBACRoles.Create(ctx, &role.RBACRole)
	if err != nil {
		return nil, err
	}
	return &state.RBACRole{RBACRole: *updatedRBACRole}, nil
}

type rbacRoleDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *rbacRoleDiffer) Deletes(handler func(crud.Event) error) error {
	currentRBACRoles, err := d.currentState.RBACRoles.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching rbac roles from state: %w", err)
	}

	for _, role := range currentRBACRoles {
		n, err := d.deleteRBACRole(role)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (d *rbacRoleDiffer) deleteRBACRole(role *state.RBACRole) (*crud.Event, error) {
	_, err := d.targetState.RBACRoles.Get(*role.Name)
	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  role,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up rbac role %q: %w",
			role.FriendlyName(), err)
	}
	return nil, nil
}

func (d *rbacRoleDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetRBACRoles, err := d.targetState.RBACRoles.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching rbac roles from state: %w", err)
	}

	for _, role := range targetRBACRoles {
		n, err := d.createUpdateRBACRole(role)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *rbacRoleDiffer) createUpdateRBACRole(role *state.RBACRole) (*crud.Event, error) {
	roleCopy := &state.RBACRole{RBACRole: *role.DeepCopy()}
	currentRole, err := d.currentState.RBACRoles.Get(*role.Name)

	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  roleCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up rbac role %q: %w",
			role.FriendlyName(), err)
	}

	// found, check if update needed
	if !currentRole.EqualWithOpts(roleCopy, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    roleCopy,
			OldObj: currentRole,
		}, nil
	}
	return nil, nil
}
