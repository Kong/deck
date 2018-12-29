package diff

import (
	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteRoutes() error {
	currentRoutes, err := sc.currentState.Routes.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching routes from state")
	}

	for _, route := range currentRoutes {
		ok, err := sc.deleteRoute(route)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		n := Node{
			Op:   crud.Delete,
			Kind: "route",
			Obj:  route,
		}
		sc.sendEvent(n)
	}
	return nil
}

func (sc *Syncer) deleteRoute(route *state.Route) (bool, error) {
	if utils.Empty(route.Name) {
		return false, errors.New("'name' attribute for a route cannot be nil")
	}
	if route.Service == nil ||
		(utils.Empty(route.Service.ID)) {
		return false, errors.Errorf("route has no associated service: %+v", route)
	}
	// lookup by Name
	_, err := sc.targetState.Routes.Get(*route.Name)
	if err == state.ErrNotFound {
		return true, nil
	}
	if err != nil {
		return false, errors.Wrapf(err, "looking up route '%v'", *route.Name)
	}
	return false, nil
}

func (sc *Syncer) createUpdateRoutes() error {
	targetRoutes, err := sc.targetState.Routes.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching routes from state")
	}

	for _, route := range targetRoutes {
		err := sc.createUpdateRoute(route)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *Syncer) createUpdateRoute(route *state.Route) error {
	route = &state.Route{Route: *route.DeepCopy()}
	currentRoute, err := sc.currentState.Routes.Get(*route.Name)
	if err == state.ErrNotFound {
		// route not present, create it

		// XXX fill foreign
		svc, err := sc.currentState.Services.Get(*route.Service.Name)
		if err != nil {
			return errors.Wrapf(err, "could not find service '%v' for route %+v", *route.Service.Name, *route.Name)
		}
		route.Service = &svc.Service
		// XXX

		route.ID = nil
		n := Node{
			Op:   crud.Create,
			Kind: "route",
			Obj:  route,
		}
		sc.sendEvent(n)
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "error looking up route %v", *route.Name)
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
			return errors.Wrapf(err, "looking up service '%v' for route '%v'", *route.Service.Name, *route.Name)
		}
		route.Service.ID = svc.ID
		// XXX
		n := Node{
			Op:     crud.Update,
			Kind:   "route",
			Obj:    route,
			OldObj: currentRoute,
		}
		sc.sendEvent(n)
	}
	return nil
}
