package cmd

import (
	"fmt"

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
	if syncNoMerge && len(applyCmdKongStateFile) > 0 {
		// Save original SelectorTags from CLI flags. syncMain mutates the global
		// dumpConfig.SelectorTags via determineSelectorTag (picking up tags from
		// each file's _info.select_tags), so we must restore them before each
		// iteration to prevent the first file's tags from bleeding into the next.
		originalSelectorTags := dumpConfig.SelectorTags
		for _, file := range applyCmdKongStateFile {
			if file == "-" {
				return fmt.Errorf("cannot use --no-merge with stdin input")
			}

			dumpConfig.SelectorTags = originalSelectorTags

			err := syncMain(cmd.Context(), []string{file}, false,
				applyCmdParallelism, applyCmdDBUpdateDelay, applyWorkspace, applyJSONOutput, ApplyTypePartial)
			if err != nil {
				return err
			}
		}

		return nil
	}

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
	applyCmd.Flags().BoolVar(&dumpConfig.SkipHashForBasicAuth, "skip-hash-for-basic-auth",
		false, "do not sync hash for basic auth credentials.\n"+
			"This flag is only valid with Konnect.")
	applyCmd.Flags().BoolVar(&syncJSONOutput, "json-output",
		false, "generate command execution report in a JSON format")
	addSilenceEventsFlag(applyCmd.Flags())
	applyCmd.Flags().BoolVar(&syncNoMerge, "no-merge",
		false, "do not merge the state file with the existing configuration")

	return applyCmd
}
