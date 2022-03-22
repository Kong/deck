package types

import (
	"context"
	"fmt"
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
	event := crud.EventFromArg(arg[0])
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
	event := crud.EventFromArg(arg[0])
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
	event := crud.EventFromArg(arg[0])
	ep := rbacEndpointPermissionFromStruct(event)

	updatedRBACEndpointPermission, err := s.client.RBACEndpointPermissions.Update(ctx, &ep.RBACEndpointPermission)
	if err != nil {
		return nil, err
	}
	return &state.RBACEndpointPermission{RBACEndpointPermission: *updatedRBACEndpointPermission}, nil
}

func (d *rbacEndpointPermissionDiffer) Deletes(handler func(crud.Event) error) error {
	currentRBACEndpointPermissions, err := d.currentState.RBACEndpointPermissions.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching eps from state: %w", err)
	}

	for _, ep := range currentRBACEndpointPermissions {
		n, err := d.deleteRBACEndpointPermission(ep)
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

type rbacEndpointPermissionDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *rbacEndpointPermissionDiffer) deleteRBACEndpointPermission(ep *state.RBACEndpointPermission) (
	*crud.Event, error,
) {
	_, err := d.targetState.RBACEndpointPermissions.Get(ep.Identifier())
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  ep,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up rbac ep %q: %w",
			ep.ID, err)
	}
	return nil, nil
}

func (d *rbacEndpointPermissionDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetRBACEndpointPermissions, err := d.targetState.RBACEndpointPermissions.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching rbac eps from state: %w", err)
	}

	for _, ep := range targetRBACEndpointPermissions {
		n, err := d.createUpdateRBACEndpointPermission(ep)
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

func (d *rbacEndpointPermissionDiffer) createUpdateRBACEndpointPermission(ep *state.RBACEndpointPermission) (
	*crud.Event, error,
) {
	epCopy := &state.RBACEndpointPermission{RBACEndpointPermission: *ep.DeepCopy()}
	currentEp, err := d.currentState.RBACEndpointPermissions.Get(ep.Identifier())

	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  epCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up rbac endpoint permission %q: %w",
			ep.Identifier(), err)
	}

	// found, check if update needed
	if !currentEp.EqualWithOpts(epCopy, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    epCopy,
			OldObj: currentEp,
		}, nil
	}
	return nil, nil
}
