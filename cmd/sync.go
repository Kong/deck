// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"errors"

	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/dump"
	"github.com/hbagdi/deck/file"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
	"github.com/spf13/cobra"
)

var syncCmdKongStateFile string

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync performs operations to get Kong's configuration to match the state file",
	Long: `Sync command reads the state file and performs operation on Kong to get Kong's
state in sync with the input state.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		current, err := state.NewKongState()
		if err != nil {
			return err
		}
		client, err := kong.NewClient(nil, nil)
		if err != nil {
			return err
		}
		services, err := dump.GetAllServices(client)
		if err != nil {
			return err
		}
		for _, service := range services {
			var s state.Service
			s.Service = *service
			err := current.AddService(s)
			if err != nil {
				return err
			}
		}
		routes, err := dump.GetAllRoutes(client)
		if err != nil {
			return err
		}
		for _, route := range routes {
			var r state.Route
			r.Route = *route
			err := current.AddRoute(r)
			if err != nil {
				return err
			}
		}
		target, err := file.GetStateFromFile(syncCmdKongStateFile)
		if err != nil {
			return err
		}
		s, _ := diff.NewSyncer(current, target)
		gDelete, gCreateUpdate, err := s.Diff()
		if err != nil {
			return err
		}
		err = s.Solve(gDelete, client)
		if err != nil {
			return err
		}
		err = s.Solve(gCreateUpdate, client)
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
