package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	syncCmdParallelism   int
	syncCmdDBUpdateDelay int
	syncWorkspace        string
	syncJSONOutput       bool
)

var syncCmdKongStateFile []string

func executeSync(cmd *cobra.Command, _ []string) error {
	return syncMain(cmd.Context(), syncCmdKongStateFile, false,
		syncCmdParallelism, syncCmdDBUpdateDelay, syncWorkspace, syncJSONOutput, ApplyTypeFull)
}

// newSyncCmd represents the sync command
func newSyncCmd(deprecated bool) *cobra.Command {
	use := "sync [flags] [kong-state-files...]"
	short := "Sync performs operations to get Kong's configuration to match the state file"
	execute := executeSync
	argsValidator := cobra.MinimumNArgs(0)
	preRun := func(_ *cobra.Command, args []string) error {
		syncCmdKongStateFile = args
		if len(syncCmdKongStateFile) == 0 {
			syncCmdKongStateFile = []string{"-"}
		}
		return preRunSilenceEventsFlag()
	}

	if deprecated {
		use = "sync"
		short = "[deprecated] see 'deck gateway sync --help' for changes to the command"
		execute = func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stderr, "Info: 'deck sync' functionality has moved to 'deck gateway sync' and will be removed\n"+
				"in a future MAJOR version of deck. Migration to 'deck gateway sync' is recommended.\n"+
				"   Note: - see 'deck gateway sync --help' for changes to the command\n"+
				"         - files changed to positional arguments without the '-s/--state' flag\n"+
				"         - the default changed from 'kong.yaml' to '-' (stdin/stdout)\n")
			return executeSync(cmd, args)
		}
		argsValidator = validateNoArgs
		preRun = func(_ *cobra.Command, _ []string) error {
			if len(syncCmdKongStateFile) == 0 {
				return fmt.Errorf("a state file with Kong's configuration " +
					"must be specified using `-s`/`--state` flag")
			}
			return preRunSilenceEventsFlag()
		}
	}

	syncCmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long: `The sync command reads the state file and performs operation on Kong
to get Kong's state in sync with the input state.`,
		Args:    argsValidator,
		RunE:    execute,
		PreRunE: preRun,
	}

	if deprecated {
		syncCmd.Flags().StringSliceVarP(&syncCmdKongStateFile,
			"state", "s", []string{"kong.yaml"}, "file(s) containing Kong's configuration.\n"+
				"This flag can be specified multiple times for multiple files.\n"+
				"Use `-` to read from stdin.")
	}
	syncCmd.Flags().StringVarP(&syncWorkspace, "workspace", "w", "",
		"Sync configuration to a specific workspace "+
			"(Kong Enterprise only).\n"+
			"This takes precedence over _workspace fields in state files.")
	syncCmd.Flags().BoolVar(&dumpConfig.SkipConsumers, "skip-consumers",
		false, "do not sync consumers, consumer-groups or "+
			"any plugins associated with them.")
	syncCmd.Flags().IntVar(&syncCmdParallelism, "parallelism",
		10, "Maximum number of concurrent operations.")
	syncCmd.Flags().BoolVar(&noMaskValues, "no-mask-deck-env-vars-value",
		false, "do not mask DECK_ environment variable values at diff output.")
	syncCmd.Flags().StringSliceVar(&dumpConfig.SelectorTags,
		"select-tag", []string{},
		"only entities matching tags specified via this flag are synced.\n"+
			"When this setting has multiple tag values, entities must match every tag.\n"+
			"All entities in the state file will get the select-tags assigned if not present already.")
	syncCmd.Flags().BoolVar(&dumpConfig.RBACResourcesOnly, "rbac-resources-only",
		false, "diff only the RBAC resources (Kong Enterprise only).")
	syncCmd.Flags().IntVar(&syncCmdDBUpdateDelay, "db-update-propagation-delay",
		0, "artificial delay (in seconds) that is injected between insert operations \n"+
			"for related entities (usually for Cassandra deployments).\n"+
			"See `db_update_propagation` in kong.conf.")
	syncCmd.Flags().BoolVar(&dumpConfig.SkipCACerts, "skip-ca-certificates",
		false, "do not sync CA certificates.")
	syncCmd.Flags().BoolVar(&dumpConfig.IsConsumerGroupPolicyOverrideSet, "consumer-group-policy-overrides",
		false, "allow deck to sync consumer-group policy overrides.\n"+
			"This allows policy overrides to work with Kong GW versions >= 3.4\n"+
			"Warning: do not mix with consumer-group scoped plugins")
	syncCmd.Flags().BoolVar(&dumpConfig.SkipConsumersWithConsumerGroups, "skip-consumers-with-consumer-groups",
		false, "do not show the association between consumer and consumer-group.\n"+
			"If set to true, deck skips listing consumers with consumer-groups,\n"+
			"thus gaining some performance with large configs.\n"+
			"Usage of this flag without apt select-tags and default-lookup-tags can be problematic.\n"+
			"This flag is not valid with Konnect.")
	syncCmd.Flags().BoolVar(&syncCmdAssumeYes, "yes",
		false, "assume `yes` to prompts and run non-interactively.")
	syncCmd.Flags().BoolVar(&syncJSONOutput, "json-output",
		false, "generate command execution report in a JSON format")
	addSilenceEventsFlag(syncCmd.Flags())
	return syncCmd
}
