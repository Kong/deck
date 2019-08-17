// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/dump"
	"github.com/hbagdi/deck/file"
	"github.com/hbagdi/deck/solver"
	"github.com/hbagdi/deck/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var syncCmdKongStateFile string

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use: "sync",
	Short: "Sync performs operations to get Kong's configuration " +
		"to match the state file",
	Long: `Sync command reads the state file and performs operation on Kong
to get Kong's state in sync with the input state.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetState, selectTags, workspace, err :=
			file.GetStateFromFile(syncCmdKongStateFile)
		if err != nil {
			return err
		}
		config.Workspace = workspace
		client, err := utils.GetKongClient(config)
		if err != nil {
			return err
		}
		dumpConfig.SelectorTags = selectTags
		currentState, err := dump.GetState(client, dumpConfig)
		if err != nil {
			return err
		}
		syncer, _ := diff.NewSyncer(currentState, targetState)
		errs := solver.Solve(stopChannel, syncer, client, false)
		if errs != nil {
			return utils.ErrArray{Errors: errs}
		}
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if syncCmdKongStateFile == "" {
			return errors.New("A state file with Kong's configuration " +
				"must be specified using -s/--state flag.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringVarP(&syncCmdKongStateFile,
		"state", "s", "kong.yaml", "file containing Kong's configuration. "+
			"Use '-' to read from stdin.")
	syncCmd.Flags().BoolVar(&dumpConfig.SkipConsumers, "skip-consumers",
		false, "do not diff consumers or "+
			"any plugins associated with consumers")
}
