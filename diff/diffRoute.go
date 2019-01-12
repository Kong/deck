package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
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
	if utils.Empty(route.Name) {
		return nil, errors.New("'name' attribute for a route cannot be nil")
	}
	if route.Service == nil ||
		(utils.Empty(route.Service.ID)) {
		return nil, errors.Errorf("route has no associated service: %+v", route)
	}
	// lookup by Name
	_, err := sc.targetState.Routes.Get(*route.Name)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "route",
			Obj:  route,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up route '%v'", *route.Name)
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
	currentRoute, err := sc.currentState.Routes.Get(*route.Name)
	if err == state.ErrNotFound {
		// route not present, create it

		// XXX fill foreign
		svc, err := sc.currentState.Services.Get(*route.Service.Name)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not find service '%v' for route %+v",
				*route.Service.Name, *route.Name)
		}
		route.Service = &svc.Service
		// XXX

		route.ID = nil
		return &Event{
			Op:   crud.Create,
			Kind: "route",
			Obj:  route,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up route %v", *route.Name)
	}
	currentRoute = &state.Route{Route: *currentRoute.DeepCopy()}
	// found, check if update needed

	currentRoute.Service = &kong.Service{Name: currentRoute.Service.Name}
	route.Service = &kong.Service{Name: route.Service.Name}
	if !currentRoute.EqualWithOpts(route, true, true, false) {
		route.ID = kong.String(*currentRoute.ID)

		// XXX fill foreign
		svc, err := sc.currentState.Services.Get(*route.Service.Name)
		if err != nil {
			return nil, errors.Wrapf(err,
				"looking up service '%v' for route '%v'",
				*route.Service.Name, *route.Name)
		}
		route.Service.ID = svc.ID
		// XXX
		return &Event{
			Op:     crud.Update,
			Kind:   "route",
			Obj:    route,
			OldObj: currentRoute,
		}, nil
	}
	return nil, nil
}
