package diff

import (
	"fmt"

	"github.com/hashicorp/terraform/dag"
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteRoutes() error {
	currentRoutes, err := sc.currentState.GetAllRoutes()
	if err != nil {
		return errors.Wrap(err, "error fetching routes from state")
	}

	for _, route := range currentRoutes {
		fmt.Println("            ")
		fmt.Println("considering for delete", route)
		_, err := sc.deleteRoute(route)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *Syncer) deleteRoute(route *state.Route) (bool, error) {

	if route.Service == nil ||
		(utils.Empty(route.Service.ID) && utils.Empty(route.Service.Name)) {
		return false, errors.Errorf("route has no associated service: %+v", route)
	}
	service, err := sc.currentState.GetService(*route.Service.ID)
	if err != nil {
		return false, errors.Wrap(err, "no service found with ID "+*route.Service.ID)
	}
	node := service.Meta.GetMeta(nodeKey)
	if node != nil {
		// delete this node if the service is to be deleted
		serviceGraphNode := node.(*Node)
		if serviceGraphNode.Op == crud.Delete {
			fmt.Println("deleting route ", *route.ID)
			n := &Node{
				Op:   crud.Delete,
				Kind: "route",
				Obj:  route,
			}
			sc.deleteGraph.Add(n)
			// Route needs to be deleted before service
			sc.deleteGraph.Connect(dag.BasicEdge(serviceGraphNode, n))
			return true, nil
		}
	}
	// lookup the route by ID
	r, err := sc.targetState.GetRoute(*route.ID)
	if err == nil && r != nil {
		return false, nil
	}
	// TODO add lookup by name post Kong 1.0

	// Try best effort matching
	routes, err := sc.targetState.GetAllRoutesByServiceName(*service.Name)
	if err == state.ErrNotFound {
		return true, nil
	}

	fmt.Println("routes", routes)
	for _, r := range routes {
		fmt.Println("comparing with: ", r)
		// if we are matching up then assign the IP of the route in
		// current state to target state so that it matches things correctly
		if !r.EqualWithOpts(route, true, true, true) {
			continue
		}
		fmt.Println("equal routes found")
		if r.ID != nil && *r.ID == *route.ID {
			return false, nil
		}
		if isPlaceHolder(r.ID) {
			fmt.Println("placeholder route found")
			r.ID = kong.String(*route.ID)
			err = sc.currentState.UpdateRoute(*r)
			return false, nil
		}
	}
	fmt.Println("deleting route ", *route.ID)
	n := &Node{
		Op:   crud.Delete,
		Kind: "route",
		Obj:  route,
	}
	sc.deleteGraph.Add(n)
	return true, nil
}

func (sc *Syncer) createUpdateRoutes() error {

	targetRoutes, err := sc.targetState.GetAllRoutes()
	if err != nil {
		return errors.Wrap(err, "error fetching routes from state")
	}

	for _, route := range targetRoutes {
		err := sc.createUpdateRoute(route)
		fmt.Println("considering for create", *route.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *Syncer) createUpdateRoute(route *state.Route) error {
	route = &state.Route{Route: *route.DeepCopy()}
	fmt.Println("route", route)
	_, err := sc.currentState.GetRoute(*route.ID)
	if err == state.ErrNotFound {
		route.ID = nil
		n := &Node{
			Op:   crud.Create,
			Kind: "route",
			Obj:  route,
		}
		sc.createUpdateGraph.Add(n)
		svc, err := sc.targetState.GetService(*route.Service.Name)
		if err != nil {
			return errors.Wrapf(err, "couldn't find service for route %+v", route)
		}
		fmt.Println("found service", svc, err)
		node := svc.Meta.GetMeta(nodeKey)
		fmt.Println("found node as well", node)
		if node != nil {
			// delete this node if the service is to be deleted
			serviceGraphNode := node.(*Node)
			if serviceGraphNode.Op == crud.Create {
				fmt.Println("adding edge")
				sc.createUpdateGraph.Connect(dag.BasicEdge(n, serviceGraphNode))
			}
		}
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "error looking up service")
	}
	// // if found, check if update needed
	// if !r.EqualWithOpts(route, true, true) {
	// 	route.ID = kong.String(*s.ID)
	// 	sc.createUpdateGraph.Add(Node{
	// 		Op:   crud.Update,
	// 		Kind: "service",
	// 		Obj:  service,
	// 	})
	// }
	return nil
}
