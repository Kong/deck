package types

import (
	"context"

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
