package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// konnectSyncCmd represents the 'deck konnect diff' command.
var konnectSyncCmd = &cobra.Command{
	Use: "sync",
	Short: "Sync performs operations to get Konnect's configuration " +
		"to match the state file (in alpha)",
	Long: `The konnect sync command reads the state file and performs operations in Konnect
to get Konnect's state in sync with the input state.` + konnectAlphaState,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if konnectDumpCmdKongStateFile == "-" {
			return fmt.Errorf("writing to stdout is not supported in Konnect mode")
		}
		_ = sendAnalytics("konnect-sync", "")
		return syncKonnect(cmd.Context(), konnectDiffCmdKongStateFile, false,
			konnectDiffCmdParallelism)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunSilenceEventsFlag()
	},
}

func init() {
	konnectCmd.AddCommand(konnectSyncCmd)
	konnectSyncCmd.Flags().StringSliceVarP(&konnectDiffCmdKongStateFile,
		"state", "s", []string{"konnect.yaml"}, "file(s) containing Konnect's configuration.\n"+
			"This flag can be specified multiple times for multiple files.")
	konnectSyncCmd.Flags().BoolVar(&konnectDumpIncludeConsumers, "include-consumers",
		false, "export consumers, associated credentials and any plugins associated "+
			"with consumers.")
	konnectSyncCmd.Flags().IntVar(&konnectDiffCmdParallelism, "parallelism",
		100, "Maximum number of concurrent operations.")
	addSilenceEventsFlag(konnectSyncCmd.Flags())
}
