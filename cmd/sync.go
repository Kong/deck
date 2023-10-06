package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	syncCmdParallelism   int
	syncCmdDBUpdateDelay int
	syncWorkspace        string
	syncJSONOutput       bool
)

// newSyncCmd represents the sync command
func newSyncCmd() *cobra.Command {
	var syncCmdKongStateFile []string
	syncCmd := &cobra.Command{
		Use: "sync",
		Short: "Sync performs operations to get Kong's configuration " +
			"to match the state file",
		Long: `The sync command reads the state file and performs operation on Kong
to get Kong's state in sync with the input state.`,
		Args: validateNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return syncMain(cmd.Context(), syncCmdKongStateFile, false,
				syncCmdParallelism, syncCmdDBUpdateDelay, syncWorkspace, syncJSONOutput)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(syncCmdKongStateFile) == 0 {
				return fmt.Errorf("a state file with Kong's configuration " +
					"must be specified using `-s`/`--state` flag")
			}
			return preRunSilenceEventsFlag()
		},
	}

	syncCmd.Flags().StringSliceVarP(&syncCmdKongStateFile,
		"state", "s", []string{"kong.yaml"}, "file(s) containing Kong's configuration.\n"+
			"This flag can be specified multiple times for multiple files.\n"+
			"Use `-` to read from stdin.")
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
			"When this setting has multiple tag values, entities must match every tag.")
	syncCmd.Flags().BoolVar(&dumpConfig.RBACResourcesOnly, "rbac-resources-only",
		false, "diff only the RBAC resources (Kong Enterprise only).")
	syncCmd.Flags().IntVar(&syncCmdDBUpdateDelay, "db-update-propagation-delay",
		0, "artificial delay (in seconds) that is injected between insert operations \n"+
			"for related entities (usually for Cassandra deployments).\n"+
			"See `db_update_propagation` in kong.conf.")
	syncCmd.Flags().BoolVar(&dumpConfig.SkipCACerts, "skip-ca-certificates",
		false, "do not sync CA certificates.")
	syncCmd.Flags().BoolVar(&syncJSONOutput, "json-output",
		false, "generate command execution report in a JSON format")
	addSilenceEventsFlag(syncCmd.Flags())
	return syncCmd
}
