package solver

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// rbacRoleCRUD implements crud.Actions interface.
type rbacRoleCRUD struct {
	client *kong.Client
}

func rbacRoleFromStuct(arg diff.Event) *state.RBACRole {
	role, ok := arg.Obj.(*state.RBACRole)
	if !ok {
		panic("unexpected type, expected *state.RBACRole")
	}

	return role
}

// Create creates a RBACRole in Kong.
// The arg should be of type diff.Event, containing the route to be created,
// else the function will panic.
// It returns a the created *state.RBACRole.
func (s *rbacRoleCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	role := routeFromStuct(event)
	createdRBACRole, err := s.client.RBACRoles.Create(nil, role)
	if err != nil {
		return nil, err
	}
	return &state.RBACRole{RBACRole: *createdRBACRole}, nil
}

// Delete deletes a RBACRole in Kong.
// The arg should be of type diff.Event, containing the route to be deleted,
// else the function will panic.
// It returns a the deleted *state.RBACRole.
func (s *rbacRoleCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	route := routeFromStuct(event)
	err := s.client.RBACRoles.Delete(nil, route.ID)
	if err != nil {
		return nil, err
	}
	return route, nil
}

// Update updates a RBACRole in Kong.
// The arg should be of type diff.Event, containing the route to be updated,
// else the function will panic.
// It returns a the updated *state.RBACRole.
func (s *rbacRoleCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	route := routeFromStuct(event)

	updatedRBACRole, err := s.client.RBACRoles.Create(nil, &route.RBACRole)
	if err != nil {
		return nil, err
	}
	return &state.RBACRole{RBACRole: *updatedRBACRole}, nil
}
