// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"errors"

	"github.com/kong/deck/diff"
	"github.com/kong/deck/dump"
	"github.com/kong/deck/file"
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
		client, err := GetKongClient(config)
		if err != nil {
			return err
		}
		currentState, err := dump.GetState(client)
		if err != nil {
			return err
		}
		targetState, err := file.GetStateFromFile(syncCmdKongStateFile)
		if err != nil {
			return err
		}
		syncer, _ := diff.NewSyncer(currentState, targetState)
		gDelete, gCreateUpdate, err := syncer.Diff()
		if err != nil {
			return err
		}
		err = syncer.Solve(gDelete, client)
		if err != nil {
			return err
		}
		err = syncer.Solve(gCreateUpdate, client)
		if err != nil {
			return err
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
		"state", "s", "", "file containing Kong's configuration.")
}
