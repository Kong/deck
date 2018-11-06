// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"fmt"

	"github.com/hbagdi/doko/state"
	"github.com/hbagdi/doko/sync"
	"github.com/hbagdi/go-kong/kong"
	"github.com/spf13/cobra"
)

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
		s, _ := sync.NewSyncer(nil, nil)
		target, err := state.NewKongState()
		if err != nil {
			return err
		}
		current, err := state.NewKongState()
		if err != nil {
			return err
		}
		var service state.Service
		// service.Name = kong.String("foo")
		service.Name = kong.String("S")
		err = current.AddService(service)
		if err != nil {
			return err
		}

		var service2 state.Service
		// service2.Name = kong.String("foo")
		service2.Name = kong.String("S")
		err = target.AddService(service2)
		if err != nil {
			return err
		}

		se, err := target.GetServiceByName("S")
		if err != nil {
			return err
		}
		fmt.Println(*se)

		svcs, err := target.GetAllServices()
		fmt.Printf("%#v %v\n", svcs[0], err)

		g, err := s.Diff(target, current)
		if err != nil {
			return err
		}
		err = s.Solve(g)
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
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
