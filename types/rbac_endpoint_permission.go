package types

import (
	"context"
	"strings"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// rbacEndpointPermissionCRUD implements crud.Actions interface.
type rbacEndpointPermissionCRUD struct {
	client *kong.Client
}

func rbacEndpointPermissionFromStruct(arg crud.Event) *state.RBACEndpointPermission {
	ep, ok := arg.Obj.(*state.RBACEndpointPermission)
	if !ok {
		panic("unexpected type, expected *state.RBACEndpointPermission")
	}

	return ep
}

// Create creates a RBACEndpointPermission in Kong.
// The arg should be of type crud.Event, containing the ep to be created,
// else the function will panic.
// It returns a the created *state.RBACEndpointPermission.
func (s *rbacEndpointPermissionCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	ep := rbacEndpointPermissionFromStruct(event)
	createdRBACEndpointPermission, err := s.client.RBACEndpointPermissions.Create(ctx, &ep.RBACEndpointPermission)
	if err != nil {
		return nil, err
	}
	return &state.RBACEndpointPermission{RBACEndpointPermission: *createdRBACEndpointPermission}, nil
}

// Delete deletes a RBACEndpointPermission in Kong.
// The arg should be of type crud.Event, containing the ep to be deleted,
// else the function will panic.
// It returns a the deleted *state.RBACEndpointPermission.
func (s *rbacEndpointPermissionCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	ep := rbacEndpointPermissionFromStruct(event)

	// for DELETE calls, the endpoint is passed in the URL only
	// including the leading slash results in a URL like
	// /rbac/roles/ROLEID/endpoints/workspace//foo/
	// Kong expects a URL like
	// /rbac/roles/ROLEID/endpoints/workspace/foo/
	// so we strip this before passing it to go-kong
	trimmed := strings.TrimLeft(*ep.Endpoint, "/")
	err := s.client.RBACEndpointPermissions.Delete(ctx, ep.Role.ID, ep.Workspace, &trimmed)
	if err != nil {
		return nil, err
	}
	return ep, nil
}

// Update updates a RBACEndpointPermission in Kong.
// The arg should be of type crud.Event, containing the ep to be updated,
// else the function will panic.
// It returns a the updated *state.RBACEndpointPermission.
func (s *rbacEndpointPermissionCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	ep := rbacEndpointPermissionFromStruct(event)

	updatedRBACEndpointPermission, err := s.client.RBACEndpointPermissions.Update(ctx, &ep.RBACEndpointPermission)
	if err != nil {
		return nil, err
	}
	return &state.RBACEndpointPermission{RBACEndpointPermission: *updatedRBACEndpointPermission}, nil
}
