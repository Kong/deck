package sync

import (
	"github.com/hashicorp/terraform/dag"
	"github.com/hbagdi/doko/crud"
	"github.com/hbagdi/doko/event"
	"github.com/hbagdi/doko/graph"
	"github.com/hbagdi/doko/state"
	"github.com/hbagdi/go-kong/kong"
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
func (s *Syncer) Or() {
	g, _ := s.Diff(nil, nil)
	s.Solve(g)
}

func (sc *Syncer) Diff(target, current *state.KongState) (*dag.AcyclicGraph, error) {
	var g dag.AcyclicGraph

	s := &graph.Service{}
	s.Name = kong.String("first")

	g.Add(event.Node{
		Op:   crud.Create,
		Obj:  s,
		Kind: "service",
	})
	return &g, nil
}
