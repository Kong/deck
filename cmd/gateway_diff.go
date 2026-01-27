package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	diffCmdKongStateFile   []string
	diffCmdParallelism     int
	diffCmdNonZeroExitCode bool
	diffWorkspace          string
	diffJSONOutput         bool
)

func executeDiff(cmd *cobra.Command, _ []string) error {
	return syncMain(cmd.Context(), diffCmdKongStateFile, true,
		diffCmdParallelism, 0, diffWorkspace, diffJSONOutput, ApplyTypeFull)
}

// newDiffCmd represents the diff command
func newDiffCmd(deprecated bool) *cobra.Command {
	use := "diff [flags] [kong-state-files...]"
	short := "Diff the current entities in Kong with the one on disks"
	execute := executeDiff
	argsValidator := cobra.MinimumNArgs(0)
	preRun := func(_ *cobra.Command, args []string) error {
		diffCmdKongStateFile = args
		if len(diffCmdKongStateFile) == 0 {
			diffCmdKongStateFile = []string{"-"}
		}
		return preRunSilenceEventsFlag()
	}

	if deprecated {
		use = "diff"
		short = "[deprecated] see 'deck gateway diff --help' for changes to the command"
		execute = func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stderr, "Info: 'deck diff' functionality has moved to 'deck gateway diff' and will be removed\n"+
				"in a future MAJOR version of deck. Migration to 'deck gateway diff' is recommended.\n"+
				"   Note: - see 'deck gateway diff --help' for changes to the command\n"+
				"         - files changed to positional arguments without the '-s/--state' flag\n"+
				"         - the default changed from 'kong.yaml' to '-' (stdin/stdout)\n")

			return executeDiff(cmd, args)
		}
		argsValidator = validateNoArgs
		preRun = func(_ *cobra.Command, _ []string) error {
			if len(diffCmdKongStateFile) == 0 {
				return fmt.Errorf("a state file with Kong's configuration " +
					"must be specified using `-s`/`--state` flag")
			}
			return preRunSilenceEventsFlag()
		}
	}

	diffCmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long: `The diff command is similar to a dry run of the 'decK kong sync' command.

It loads entities from Kong and performs a diff with
the entities in local files. This allows you to see the entities
that will be created, updated, or deleted.
`,
		Args:    argsValidator,
		RunE:    execute,
		PreRunE: preRun,
	}

	if deprecated {
		diffCmd.Flags().StringSliceVarP(&diffCmdKongStateFile,
			"state", "s", []string{"kong.yaml"}, "file(s) containing Kong's configuration.\n"+
				"This flag can be specified multiple times for multiple files.\n"+
				"Use `-` to read from stdin.")
	}
	diffCmd.Flags().StringVarP(&diffWorkspace, "workspace", "w",
		"", "Diff configuration with a specific workspace "+
			"(Kong Enterprise only).\n"+
			"This takes precedence over _workspace fields in state files.")
	diffCmd.Flags().BoolVar(&dumpConfig.SkipConsumers, "skip-consumers",
		false, "do not diff consumers or "+
			"any plugins associated with consumers")
	diffCmd.Flags().IntVar(&diffCmdParallelism, "parallelism",
		10, "Maximum number of concurrent operations.")
	diffCmd.Flags().BoolVar(&noMaskValues, "no-mask-deck-env-vars-value",
		false, "do not mask DECK_ environment variable values at diff output.")
	diffCmd.Flags().StringSliceVar(&dumpConfig.SelectorTags,
		"select-tag", []string{},
		"only entities matching tags specified via this flag are diffed.\n"+
			"When this setting has multiple tag values, entities must match each of them.")
	diffCmd.Flags().BoolVar(&dumpConfig.RBACResourcesOnly, "rbac-resources-only",
		false, "Sync only the RBAC resources (Kong Enterprise only).")
	diffCmd.Flags().BoolVar(&diffCmdNonZeroExitCode, "non-zero-exit-code",
		false, "Return exit code 2 if there is a diff present,\n"+
			"exit code 0 if no diff is found,\n"+
			"and exit code 1 if an error occurs.")
	diffCmd.Flags().BoolVar(&dumpConfig.SkipCACerts, "skip-ca-certificates",
		false, "do not diff CA certificates.")
	diffCmd.Flags().BoolVar(&diffJSONOutput, "json-output",
		false, "Generate a JSON change report that includes a change summary and details for each entity.")
	diffCmd.Flags().BoolVar(&dumpConfig.IsConsumerGroupPolicyOverrideSet, "consumer-group-policy-overrides",
		false, "allow deck to diff consumer-group policy overrides.\n"+
			"This allows policy overrides to work with Kong GW versions >= 3.4\n"+
			"Warning: do not mix with consumer-group scoped plugins")
	diffCmd.Flags().BoolVar(&dumpConfig.SkipConsumersWithConsumerGroups, "skip-consumers-with-consumer-groups",
		false, "do not show the association between consumer and consumer-group.\n"+
			"If set to true, deck skips listing consumers with consumer-groups,\n"+
			"thus gaining some performance with large configs.\n"+
			"Usage of this flag without apt select-tags and default-lookup-tags can be problematic.\n"+
			"This flag is not valid with Konnect.")
	diffCmd.Flags().BoolVar(&syncCmdAssumeYes, "yes",
		false, "assume `yes` to prompts and run non-interactively.")
	addSilenceEventsFlag(diffCmd.Flags())
	return diffCmd
}
