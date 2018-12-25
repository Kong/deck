// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"github.com/kong/deck/diff"
	"github.com/kong/deck/dump"
	"github.com/kong/deck/file"
	"github.com/kong/deck/solver"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
		err = solver.Solve(s, client, true)
		return err
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
