package diff

import (
	"fmt"

	"github.com/hashicorp/terraform/dag"
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteRoutes() error {
	fmt.Println("delete routes called")
	currentRoutes, err := sc.currentState.GetAllRoutes()
	if err != nil {
		return errors.Wrap(err, "error fetching routes from state")
	}

	for _, route := range currentRoutes {
		fmt.Println("route in ")
		ok, err := sc.deleteRoute(route)
		if err != nil {
			return err
		}
		if ok {
			// n := &Node{
			// 	Op:   crud.Delete,
			// 	Kind: "route",
			// 	Obj:  route,
			// }
			// sc.deleteGraph.Add(n)
		}
	}
	return nil
}

func (sc *Syncer) deleteRoute(route *state.Route) (bool, error) {
	fmt.Println("considering: " + *route.ID)
	if route.Service == nil ||
		(utils.Empty(route.Service.ID) && utils.Empty(route.Service.Name)) {
		return false, errors.Errorf("route has no associated service: %+v", route)
	}
	service, err := sc.currentState.GetService(*route.Service.ID)
	if err != nil {
		return false, errors.Wrap(err, "no service found with ID "+*route.Service.ID)
	}
	serviceGraphNode := service.Meta.GetMeta(nodeKey).(*Node)
	if serviceGraphNode.Op == crud.Delete {
		// delete this node if the service is to be deleted
		n := &Node{
			Op:   crud.Delete,
			Kind: "route",
			Obj:  route,
		}
		sc.deleteGraph.Add(n)
		sc.deleteGraph.Connect(dag.BasicEdge(serviceGraphNode, n))
		return true, nil
	}
	// lookup the route by ID
	r, err := sc.targetState.GetRoute(*route.ID)
	if err == nil && r != nil {
		return false, nil
	}
	// TODO add lookup by name post Kong 1.0

	routes, err := sc.targetState.GetAllRoutesByServiceName(*service.Name)
	if err == state.ErrNotFound {
		return true, nil
	}
	for _, r := range routes {
		if r.EqualWithOpts(route, true, true) {
			return false, nil
		}
	}
	return true, nil
}
func (sc *Syncer) createUpdateRoutes() error {

	targetRoutes, err := sc.targetState.GetAllRoutes()
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
	_, err := sc.currentState.GetRoute(*route.ID)
	if err == state.ErrNotFound {
		route.ID = nil
		sc.createUpdateGraph.Add(Node{
			Op:   crud.Create,
			Kind: "route",
			Obj:  route,
		})
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
