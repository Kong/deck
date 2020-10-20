package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	syncCmdKongStateFile []string
	syncCmdParallelism   int
	syncCmdDBUpdateDelay int
	syncCmdRetries       int
	syncCmdRetryDelay    int
	syncWorkspace        string
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use: "sync",
	Short: "Sync performs operations to get Kong's configuration " +
		"to match the state file",
	Long: `Sync command reads the state file and performs operation on Kong
to get Kong's state in sync with the input state.`,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return syncMain(syncCmdKongStateFile, false, syncCmdParallelism,
			syncCmdDBUpdateDelay, syncCmdRetries, syncCmdRetryDelay, syncWorkspace)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(syncCmdKongStateFile) == 0 {
			return errors.New("A state file with Kong's configuration " +
				"must be specified using -s/--state flag.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringSliceVarP(&syncCmdKongStateFile,
		"state", "s", []string{"kong.yaml"}, "file(s) containing Kong's configuration.\n"+
			"This flag can be specified multiple times for multiple files.\n"+
			"Use '-' to read from stdin.")
	syncCmd.Flags().StringVar(&syncWorkspace, "workspace", "",
		"Sync configuration to a specific workspace "+
			"(Kong Enterprise only).\n"+
			"This takes precedence over _workspace fields in state files.")
	syncCmd.Flags().BoolVar(&dumpConfig.SkipConsumers, "skip-consumers",
		false, "do not diff consumers or "+
			"any plugins associated with consumers")
	syncCmd.Flags().IntVar(&syncCmdParallelism, "parallelism",
		10, "Maximum number of concurrent operations")
	syncCmd.Flags().StringSliceVar(&dumpConfig.SelectorTags,
		"select-tag", []string{},
		"only entities matching tags specified via this flag are synced.\n"+
			"Multiple tags are ANDed together.")
	syncCmd.Flags().IntVar(&syncCmdDBUpdateDelay, "db-update-propagation-delay",
		0, "aritificial delay in seconds that is injected between insert operations \n"+
			"for related entities (usually for cassandra deployments).\n"+
			"See 'db_update_propagation' in kong.conf.")
	syncCmd.Flags().IntVar(&syncCmdRetries, "retries",
		1, "Number of times decK will retry an insert or update operation")
	syncCmd.Flags().IntVar(&syncCmdRetryDelay, "retry-delay",
		0, "Number of seconds decK will wait before retrying an insert or update operation")
}
