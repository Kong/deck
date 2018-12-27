package diff

import (
	"github.com/hashicorp/terraform/dag"
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
		_, err := sc.deleteRoute(route)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *Syncer) deleteRoute(route *state.Route) (bool, error) {
	if utils.Empty(route.Name) {
		return false, errors.New("'name' attribute for a route cannot be nil")
	}
	if route.Service == nil ||
		(utils.Empty(route.Service.ID) && utils.Empty(route.Service.Name)) {
		return false, errors.Errorf("route has no associated service: %+v", route)
	}
	deleteRoute := false
	// If parent entity is being deleted, delete this as well
	service, err := sc.currentState.Services.Get(*route.Service.ID)
	if err != nil {
		return false, errors.Wrap(err, "no service found with ID "+*route.Service.ID)
	}
	node := service.Meta.GetMeta(nodeKey)
	if node != nil {
		// delete this node if the service is to be deleted
		serviceGraphNode := node.(*Node)
		if serviceGraphNode.Op == crud.Delete {
			deleteRoute = true
		}
	}
	// lookup by Name
	_, err = sc.targetState.Routes.Get(*route.Name)
	if err == state.ErrNotFound {
		deleteRoute = true
	} else {
		return false, errors.Wrapf(err, "looking up route '%v'", *route.Name)
	}
	if deleteRoute {
		n := &Node{
			Op:   crud.Delete,
			Kind: "route",
			Obj:  route,
		}
		sc.deleteGraph.Add(n)
		route.AddMeta(nodeKey, n)
		sc.currentState.Routes.Update(*route)
		return true, nil
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
	routeCopy := &state.Route{Route: *route.DeepCopy()}
	// route should be created or updated

	// search
	currentRoute, err := sc.currentState.Routes.Get(*route.Name)
	if err == state.ErrNotFound {
		// create it
		routeCopy := &state.Route{Route: *route.DeepCopy()}

		svc, err := sc.targetState.Services.Get(*route.Service.Name)
		if err != nil {
			return errors.Wrapf(err, "couldn't find service for route %+v", route)
		}
		routeCopy.ID = nil
		n := &Node{
			Op:   crud.Create,
			Kind: "route",
			Obj:  routeCopy,
		}
		sc.createUpdateGraph.Add(n)

		node := svc.Meta.GetMeta(nodeKey)
		if node != nil {
			// foreign service needs to be created before this route can be created
			serviceGraphNode := node.(*Node)
			if serviceGraphNode.Op == crud.Create {
				sc.createUpdateGraph.Connect(dag.BasicEdge(n, serviceGraphNode))
			}
		}
		route.AddMeta(nodeKey, n)
		sc.targetState.Routes.Update(*route)
		return nil
	}
	// if found, check if update needed

	// TODO if the new service is being created, then add dependency
	// first fill foreign keys; those could have changed
	currentRouteCopy := &state.Route{Route: *currentRoute.DeepCopy()}
	if err != nil {
		return errors.Wrap(err, "error looking up route")
	}

	routeCopy.Service = &kong.Service{Name: kong.String(*route.Service.Name)}
	svcForCurrentRoute, err := sc.currentState.Services.Get(*currentRoute.Service.ID)
	if err != nil {
		return errors.Wrapf(err, "error looking up service for route '%v'", *currentRoute.ID)
	}
	currentRouteCopy.Service = &kong.Service{Name: svcForCurrentRoute.Name}

	if !currentRouteCopy.EqualWithOpts(routeCopy, true, true, false) {
		routeCopy.ID = kong.String(*currentRoute.ID)
		n := &Node{
			Op:     crud.Update,
			Kind:   "route",
			Obj:    routeCopy,
			OldObj: currentRouteCopy,
		}
		sc.createUpdateGraph.Add(n)
		route.AddMeta(nodeKey, n)
		sc.targetState.Routes.Update(*route)
	}
	return nil
}
