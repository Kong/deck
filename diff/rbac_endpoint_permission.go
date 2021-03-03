package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteRBACEndpointPermissions() error {
	currentRBACEndpointPermissions, err := sc.currentState.RBACEndpointPermissions.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching eps from state")
	}

	for _, ep := range currentRBACEndpointPermissions {
		n, err := sc.deleteRBACEndpointPermission(ep)
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

func (sc *Syncer) deleteRBACEndpointPermission(ep *state.RBACEndpointPermission) (*Event, error) {
	_, err := sc.targetState.RBACEndpointPermissions.Get(ep.Identifier())
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "rbac-endpointpermission",
			Obj:  ep,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up rbac ep '%v'",
			ep.ID)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateRBACEndpointPermissions() error {
	targetRBACEndpointPermissions, err := sc.targetState.RBACEndpointPermissions.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching rbac eps from state")
	}

	for _, ep := range targetRBACEndpointPermissions {
		n, err := sc.createUpdateRBACEndpointPermission(ep)
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

func (sc *Syncer) createUpdateRBACEndpointPermission(ep *state.RBACEndpointPermission) (*Event, error) {
	epCopy := &state.RBACEndpointPermission{RBACEndpointPermission: *ep.DeepCopy()}
	currentEp, err := sc.currentState.RBACEndpointPermissions.Get(ep.Identifier())

	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Create,
			Kind: "rbac-endpointpermission",
			Obj:  epCopy,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up rbac endpoint permission %v",
			ep.Identifier())
	}

	// found, check if update needed
	if !currentEp.EqualWithOpts(epCopy, false, true, false) {
		return &Event{
			Op:     crud.Update,
			Kind:   "rbac-endpointpermission",
			Obj:    epCopy,
			OldObj: currentEp,
		}, nil
	}
	return nil, nil
}
