package sync

import (
	"github.com/hashicorp/terraform/dag"
	"github.com/hbagdi/doko/crud"
	"github.com/hbagdi/doko/state"
	"github.com/hbagdi/doko/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

type Syncer struct {
	currentState, targetState      *state.KongState
	deleteGraph, createUpdateGraph *dag.AcyclicGraph
	registry                       crud.Registry
}

func NewSyncer(current, target *state.KongState) (*Syncer, error) {
	s := &Syncer{}
	s.currentState, s.targetState = current, target
	s.deleteGraph = new(dag.AcyclicGraph)
	s.createUpdateGraph = new(dag.AcyclicGraph)
	s.registry.Register("service", &ServiceCRUD{})
	return s, nil
}

func (sc *Syncer) Diff() (*dag.AcyclicGraph, *dag.AcyclicGraph, error) {

	err := sc.delete()
	if err != nil {
		return nil, nil, errors.Wrap(err, "couldn't create graph")
	}
	err = sc.createUpdate()
	if err != nil {
		return nil, nil, errors.Wrap(err, "couldn't create graph")
	}
	return sc.deleteGraph, sc.createUpdateGraph, nil
}

func (sc *Syncer) delete() error {
	err := sc.deleteServices()
	// deleteServices(g,target,current)
	if err != nil {
		return errors.Wrap(err, "while building graph ")
	}
	return nil
}

func (sc *Syncer) deleteServices() error {
	currentServices, err := sc.currentState.GetAllServices()
	if err != nil {
		return errors.Wrap(err, "error fetching services from state")
	}

	for _, service := range currentServices {
		ok, err := sc.deleteService(service)
		if err != nil {
			return err
		}
		if ok {
			sc.deleteGraph.Add(Node{
				Op:   crud.Delete,
				Kind: "service",
				Obj:  service,
			})
		}
	}
	return nil
}

func (sc *Syncer) deleteService(service *state.Service) (bool, error) {
	// lookup by name
	if utils.Empty(service.Name) {
		return false, errors.New("'name' attribute for a service cannot be nil")
	}
	_, err := sc.targetState.GetService(*service.Name)
	if err == state.ErrNotFound {
		return true, nil
	}
	// any other type of error
	if err != nil {
		return false, err
	}
	return false, nil
}

func (sc *Syncer) createUpdate() error {
	err := sc.createUpdateServices()
	if err != nil {
		return errors.Wrap(err, "while building graph")
	}
	return nil
}

func (sc *Syncer) createUpdateServices() error {

	targetServices, err := sc.targetState.GetAllServices()
	if err != nil {
		return errors.Wrap(err, "error fetching services from state")
	}

	for _, service := range targetServices {
		err := sc.createUpdateService(service)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *Syncer) createUpdateService(service *state.Service) error {
	service = &state.Service{*service.DeepCopy()}
	s, err := sc.currentState.GetService(*service.Name)
	if err == state.ErrNotFound {
		service.ID = nil
		sc.createUpdateGraph.Add(Node{
			Op:   crud.Create,
			Kind: "service",
			Obj:  service,
		})
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "error looking up service")
	}
	// if found, check if update needed
	if !s.EqualWithOpts(service, true, true) {
		service.ID = kong.String(*s.ID)
		sc.createUpdateGraph.Add(Node{
			Op:   crud.Update,
			Kind: "service",
			Obj:  service,
		})
	}
	return nil
}

func (s *Syncer) Solve(g *dag.AcyclicGraph, client *kong.Client) error {
	err := g.Walk(func(v dag.Vertex) error {
		n, ok := v.(Node)
		if !ok {
			panic("unexpected type encountered while solving the graph")
		}
		// every Node will need to add a few things to arg:
		// *kong.Client to use
		// callbacks to execute
		s.registry.Do(n.Kind, n.Op, n.Obj, s.currentState, s.targetState, client)
		return nil
	})
	return err
}
