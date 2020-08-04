package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteRoutes() error {
	currentRoutes, err := sc.currentState.Routes.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching routes from state")
	}

	for _, route := range currentRoutes {
		n, err := sc.deleteRoute(route)
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

func (sc *Syncer) deleteRoute(route *state.Route) (*Event, error) {
	_, err := sc.targetState.Routes.Get(*route.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "route",
			Obj:  route,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up route '%v'",
			route.Identifier())
	}
	return nil, nil
}

func (sc *Syncer) createUpdateRoutes() error {
	targetRoutes, err := sc.targetState.Routes.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching routes from state")
	}

	for _, route := range targetRoutes {
		n, err := sc.createUpdateRoute(route)
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

func (sc *Syncer) createUpdateRoute(route *state.Route) (*Event, error) {
	route = &state.Route{Route: *route.DeepCopy()}
	currentRoute, err := sc.currentState.Routes.Get(*route.ID)
	if err == state.ErrNotFound {
		// route not present, create it

		return &Event{
			Op:   crud.Create,
			Kind: "route",
			Obj:  route,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up route %v",
			route.Identifier())
	}
	// found, check if update needed

	if !currentRoute.EqualWithOpts(route, false, true, false) {
		return &Event{
			Op:     crud.Update,
			Kind:   "route",
			Obj:    route,
			OldObj: currentRoute,
		}, nil
	}
	return nil, nil
}
