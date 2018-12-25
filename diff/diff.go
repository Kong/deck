package diff

import (
	"github.com/hashicorp/terraform/dag"
	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/crud"
	cruds "github.com/kong/deck/crud/kong"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

// Syncer takes in a current and target state of Kong,
// diffs them, generating a Graph to get Kong from current
// to target state.
type Syncer struct {
	currentState, targetState      *state.KongState
	deleteGraph, createUpdateGraph *dag.AcyclicGraph
	registry                       crud.Registry
}

// NewSyncer constructs a Syncer.
func NewSyncer(current, target *state.KongState) (*Syncer, error) {
	s := &Syncer{}
	s.currentState, s.targetState = current, target
	s.deleteGraph = new(dag.AcyclicGraph)
	s.createUpdateGraph = new(dag.AcyclicGraph)
	s.registry.Register("service", &cruds.ServiceCRUD{})
	s.registry.Register("route", &cruds.RouteCRUD{})
	return s, nil
}

// Diff diffs the current and target states and returns two graphs.
// The first graph contains all the entities which should be deleted from Kong
// and the second graph contains the entities which should be created or
// updated to get Kong to target state.
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
	// TODO write an interface and register by types,
	// then execute in a particular order
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

// Solve walks a graph and executes actions.
func (sc *Syncer) Solve(g *dag.AcyclicGraph, client *kong.Client) error {
	err := g.Walk(func(v dag.Vertex) error {
		n, ok := v.(*Node)
		if !ok {
			panic("unexpected type encountered while solving the graph")
		}
		// every Node will need to add a few things to arg:
		// *kong.Client to use
		// callbacks to execute
		_, err := sc.registry.Do(n.Kind, n.Op, cruds.ArgStruct{
			Obj:    n.Obj,
			OldObj: n.OldObj,

			CurrentState: sc.currentState,
			TargetState:  sc.targetState,

			Client: client,
		})
		return err
	})
	return err
}
