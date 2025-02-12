package cmd

import (
	"github.com/spf13/cobra"
)

var (
	applyCmdParallelism   int
	applyCmdDBUpdateDelay int
	applyWorkspace        string
	applyJSONOutput       bool
)

var applyCmdKongStateFile []string

func executeApply(cmd *cobra.Command, _ []string) error {
	return syncMain(cmd.Context(), applyCmdKongStateFile, false,
		applyCmdParallelism, applyCmdDBUpdateDelay, applyWorkspace, applyJSONOutput, ApplyTypePartial)
}

func newApplyCmd() *cobra.Command {
	short := "Apply configuration to Kong without deleting existing entities"
	execute := executeApply

	applyCmd := &cobra.Command{
		Use:   "apply [flags] [kong-state-files...]",
		Short: short,
		Long:  `The apply command allows you to apply partial Kong configuration files without deleting existing entities.`,
		Args:  cobra.MinimumNArgs(0),
		RunE:  execute,
		PreRunE: func(_ *cobra.Command, args []string) error {
			applyCmdKongStateFile = args
			if len(applyCmdKongStateFile) == 0 {
				applyCmdKongStateFile = []string{"-"}
			}
			return preRunSilenceEventsFlag()
		},
	}

	applyCmd.Flags().StringVarP(&applyWorkspace, "workspace", "w", "",
		"Apply configuration to a specific workspace "+
			"(Kong Enterprise only).\n"+
			"This takes precedence over _workspace fields in state files.")
	applyCmd.Flags().IntVar(&applyCmdParallelism, "parallelism",
		10, "Maximum number of concurrent operations.")
	applyCmd.Flags().IntVar(&applyCmdDBUpdateDelay, "db-update-propagation-delay",
		0, "artificial delay (in seconds) that is injected between insert operations \n"+
			"for related entities (usually for Cassandra deployments).\n"+
			"See `db_update_propagation` in kong.conf.")
	applyCmd.Flags().BoolVar(&syncJSONOutput, "json-output",
		false, "generate command execution report in a JSON format")
	addSilenceEventsFlag(applyCmd.Flags())

	return applyCmd
}
