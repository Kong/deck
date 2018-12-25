// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"github.com/kong/deck/diff"
	"github.com/kong/deck/dump"
	"github.com/kong/deck/file"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/hashicorp/terraform/dag"
	"github.com/kong/deck/crud"
	cruds "github.com/kong/deck/crud/kong"
	drycrud "github.com/kong/deck/crud/kong/dry"
)

var diffCmdKongStateFile string

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines
and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := GetKongClient(config)
		if err != nil {
			return err
		}
		currentState, err := dump.GetState(client)
		if err != nil {
			return err
		}
		targetState, err := file.GetStateFromFile(diffCmdKongStateFile)
		if err != nil {
			return err
		}
		s, _ := diff.NewSyncer(currentState, targetState)
		gDelete, gCreateUpdate, err := s.Diff()
		if err != nil {
			return err
		}
		err = Solve(gDelete)
		if err != nil {
			return err
		}
		err = Solve(gCreateUpdate)
		if err != nil {
			return err
		}
		return nil

	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if diffCmdKongStateFile == "" {
			return errors.New("A state file with Kong's configuration " +
				"must be specified using -s/--state flag.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().StringVarP(&diffCmdKongStateFile,
		"state", "s", "", "file containing Kong's configuration.")
}

// Solve walks the graph and prints the changes out.
// This should be used for dry run of a sync graph.
// It doesn't actually make any calls to Kong's Admin API.
func Solve(g *dag.AcyclicGraph) error {
	var r crud.Registry
	r.Register("service", &drycrud.ServiceCRUD{})
	r.Register("route", &drycrud.RouteCRUD{})
	err := g.Walk(func(v dag.Vertex) error {
		n, ok := v.(*diff.Node)
		if !ok {
			panic("unexpected type encountered while solving the graph")
		}
		// every Node will need to add a few things to arg:
		// *kong.Client to use
		// callbacks to execute
		_, err := r.Do(n.Kind, n.Op, cruds.ArgStruct{
			Obj:    n.Obj,
			OldObj: n.OldObj,
			// TODO inject these
			// CurrentState: sc.currentState,
			// TargetState: sc.targetState,
		})
		return err
	})
	return err
}
