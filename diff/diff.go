package diff

import (
	"github.com/hashicorp/terraform/dag"
	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	cruds "github.com/kong/deck/state/crud"
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
	s.registry.Register("service", &cruds.ServiceCRUD{})
	s.registry.Register("route", &cruds.RouteCRUD{})
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
	if err != nil {
		return errors.Wrap(err, "while building graph ")
	}
	err = sc.deleteRoutes()
	if err != nil {
		return errors.Wrap(err, "while building graph ")
	}
	return nil
}

func (sc *Syncer) createUpdate() error {
	// TODO write an interface and register by types, then execute in a particular order
	err := sc.createUpdateServices()
	if err != nil {
		return errors.Wrap(err, "while building graph")
	}
	err = sc.createUpdateRoutes()
	if err != nil {
		return errors.Wrap(err, "while building graph")
	}
	return nil
}

func (s *Syncer) Solve(g *dag.AcyclicGraph, client *kong.Client) error {
	err := g.Walk(func(v dag.Vertex) error {
		n, ok := v.(*Node)
		if !ok {
			panic("unexpected type encountered while solving the graph")
		}
		// every Node will need to add a few things to arg:
		// *kong.Client to use
		// callbacks to execute
		_, err := s.registry.Do(n.Kind, n.Op, n.Obj, s.currentState, s.targetState, client)
		return err
	})
	return err
}
