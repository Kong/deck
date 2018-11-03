package sync

import (
	"fmt"

	"github.com/hashicorp/terraform/dag"
	"github.com/hbagdi/doko/event"
)

func (s *Syncer) Solve(g *dag.AcyclicGraph) error {
	err := g.Walk(func(v dag.Vertex) error {
		n, ok := v.(event.Node)
		if !ok {
			fmt.Println("shit")
		}
		fmt.Println(n)

		s.registry.Do(n.Kind, n.Op, n.Obj)
		return nil
	})
	return err
}
