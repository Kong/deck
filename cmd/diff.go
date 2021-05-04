package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	diffCmdKongStateFile   []string
	diffCmdParallelism     int
	diffCmdNonZeroExitCode bool
	diffWorkspace          string
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff the current entities in Kong with the one on disks",
	Long: `Diff is like a dry run of 'decK sync' command.

It will load entities form Kong and then perform a diff on those with
the entities present in files locally. This allows you to see the entities
that will be created or updated or deleted.
`,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return syncMain(cmd.Context(), diffCmdKongStateFile, true,
			diffCmdParallelism, 0, diffWorkspace)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(diffCmdKongStateFile) == 0 {
			return errors.New("A state file with Kong's configuration " +
				"must be specified using -s/--state flag.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().StringSliceVarP(&diffCmdKongStateFile,
		"state", "s", []string{"kong.yaml"}, "file(s) containing Kong's configuration.\n"+
			"This flag can be specified multiple times for multiple files.\n"+
			"Use '-' to read from stdin.")
	diffCmd.Flags().StringVarP(&diffWorkspace, "workspace", "w",
		"", "Diff configuration with a specific workspace "+
			"(Kong Enterprise only).\n"+
			"This takes precedence over _workspace fields in state files.")
	diffCmd.Flags().BoolVar(&dumpConfig.SkipConsumers, "skip-consumers",
		false, "do not diff consumers or "+
			"any plugins associated with consumers")
	diffCmd.Flags().IntVar(&diffCmdParallelism, "parallelism",
		10, "Maximum number of concurrent operations")
	diffCmd.Flags().StringSliceVar(&dumpConfig.SelectorTags,
		"select-tag", []string{},
		"only entities matching tags specified via this flag are diffed.\n"+
			"Multiple tags are ANDed together.")
	diffCmd.Flags().BoolVar(&dumpConfig.RBACResourcesOnly, "rbac-resources-only",
		false, "sync only the RBAC resources (Kong Enterprise only)")
	diffCmd.Flags().BoolVar(&diffCmdNonZeroExitCode, "non-zero-exit-code",
		false, "return exit code 2 if there is a diff present,\n"+
			"exit code 0 if no diff is found,\n"+
			"and exit code 1 if an error occurs.")
}
