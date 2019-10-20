// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var diffCmdKongStateFile string

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff the current entities in Kong with the on on disks",
	Long: `Diff is like a dry run of 'decK sync' command.

It will load entities form Kong and then perform a diff on those with
the entities present in files locally. This allows you to see the entities
that will be created or updated or deleted.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sync(diffCmdKongStateFile, true)
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
		"state", "s", "kong.yaml", "file containing Kong's configuration. "+
			"Use '-' to read from stdin.")
	diffCmd.Flags().BoolVar(&dumpConfig.SkipConsumers, "skip-consumers",
		false, "do not diff consumers or "+
			"any plugins associated with consumers")
}
