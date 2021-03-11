package cmd

import (
	"github.com/spf13/cobra"
)

// konnectSyncCmd represents the 'deck konnect diff' command.
var konnectSyncCmd = &cobra.Command{
	Use: "sync",
	Short: "Sync performs operations to get Konnect's configuration " +
		"to match the state file (in alpha)",
	Long: `Sync command reads the state file and performs operation in Konnect
to get Konnect's state in sync with the input state.` + konnectAlphaState,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return syncKonnect(cmd.Context(), konnectDiffCmdKongStateFile, false,
			konnectDiffCmdParallelism)
	},
}

func init() {
	konnectCmd.AddCommand(konnectSyncCmd)
	konnectSyncCmd.Flags().StringSliceVarP(&konnectDiffCmdKongStateFile,
		"state", "s", []string{"konnect.yaml"}, "file(s) containing Konnect's configuration.\n"+
			"This flag can be specified multiple times for multiple files.\n"+
			"Use '-' to read from stdin.")
	konnectSyncCmd.Flags().BoolVar(&konnectDumpIncludeConsumers, "include-consumers",
		false, "export consumers, associated credentials and any plugins associated "+
			"with consumers")
	konnectSyncCmd.Flags().IntVar(&konnectDiffCmdParallelism, "parallelism",
		100, "Maximum number of concurrent operations")
}
