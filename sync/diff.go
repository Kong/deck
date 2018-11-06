package sync

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/dag"
	"github.com/hbagdi/doko/crud"
	"github.com/hbagdi/doko/state"
	"github.com/pkg/errors"
)

type Syncer struct {
	currentState, targetState *state.KongState
	registry                  crud.Registry
}

func NewSyncer(current, target *state.KongState) (*Syncer, error) {
	s := &Syncer{}
	s.currentState, s.targetState = current, target
	s.registry.Register("service", &ServiceCRUD{})
	return s, nil
}

func (sc *Syncer) Diff(target, current *state.KongState) (*dag.AcyclicGraph, error) {
	var g dag.AcyclicGraph

	// every Node will need to add a few things to arg:
	// *kong.Client to use
	// callbacks to execute

	err := sc.delete(&g, target, current)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't create graph")
	}
	// figure out how to make all create/updates depened on deletes or
	// have two graphs
	// figure out how to setup dependencies
	err = sc.createAndUpdate(&g, target, current)

	if err != nil {
		return nil, errors.Wrap(err, "couldn't create graph")
	}
	return &g, nil
}

func (sc *Syncer) createAndUpdate(g *dag.AcyclicGraph, target, current *state.KongState) error {

	targetServices, err := target.GetAllServices()
	if err != nil {
		return errors.Wrap(err, "error fetching services from state")
	}

	for _, service := range targetServices {
		log.Println("service in createAndUpdate: ", service)
		s, err := current.GetService(*service.ID)
		fmt.Println(s, err)
	}

	return nil
}

func (sc *Syncer) delete(g *dag.AcyclicGraph, target, current *state.KongState) error {

	currentServices, err := current.GetAllServices()
	if err != nil {
		return errors.Wrap(err, "error fetching services from state")
	}

	for _, service := range currentServices {
		s, err := target.GetService(*service.ID)
		fmt.Println(s, err)
		if err == nil {
			continue
		} else if err == state.ErrNotFound {
			// search by name
			// figure out what to do when IDs don't match up
			_, err = target.GetService(*service.Name)
			if err == nil {
				continue
			} else if err == state.ErrNotFound {
				// delete the service
				g.Add(Node{
					Op:   crud.Delete,
					Kind: "service",
					Obj:  service,
				})
			} else {
				return errors.Wrap(err, "error looking up service")
			}
		} else {
			return errors.Wrap(err, "error looking up service")
		}
	}

	return nil
}

func (s *Syncer) Solve(g *dag.AcyclicGraph) error {
	err := g.Walk(func(v dag.Vertex) error {
		n, ok := v.(Node)
		if !ok {
			fmt.Println("shit")
		}
		fmt.Println(n)

		s.registry.Do(n.Kind, n.Op, n.Obj)
		return nil
	})
	return err
}
