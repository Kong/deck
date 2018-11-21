// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/dump"
	"github.com/hbagdi/deck/file"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
	"github.com/spf13/cobra"
)

var kongStateFile string

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		target, err := state.NewKongState()
		if err != nil {
			return err
		}
		current, err := state.NewKongState()
		if err != nil {
			return err
		}
		client, err := kong.NewClient(nil, nil)
		if err != nil {
			return err
		}
		// client.SetDebugMode(true)
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
		target, err = file.GetStateFromFile(kongStateFile)
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
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	syncCmd.Flags().StringVarP(&kongStateFile, "state", "s", "", "Kong configuration directory or file")
}
