package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	konnectDiffCmdKongStateFile   []string
	konnectDiffCmdParallelism     int
	konnectDiffCmdNonZeroExitCode bool
)

// konnectDiffCmd represents the 'deck konnect diff' command.
var konnectDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff the current entities in Konnect with the one on disks (in alpha)",
	Long: `Diff is like a dry run of 'decK sync' command.

It will load entities form Konnect and then perform a diff on those with
the entities present in files locally. This allows you to see the entities
that will be created or updated or deleted.` + konnectAlphaState,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if konnectDumpCmdKongStateFile == "-" {
			return fmt.Errorf("writing to stdout is not supported in Konnect mode")
		}
		_ = sendAnalytics("konnect-diff", "")
		return syncKonnect(cmd.Context(), konnectDiffCmdKongStateFile, true,
			konnectDiffCmdParallelism)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunSilenceEventsFlag()
	},
}

func init() {
	konnectCmd.AddCommand(konnectDiffCmd)
	konnectDiffCmd.Flags().StringSliceVarP(&konnectDiffCmdKongStateFile,
		"state", "s", []string{"konnect.yaml"}, "file(s) containing Konnect's configuration.\n"+
			"This flag can be specified multiple times for multiple files.")
	konnectDiffCmd.Flags().BoolVar(&konnectDumpIncludeConsumers, "include-consumers",
		false, "export consumers, associated credentials and any plugins associated "+
			"with consumers")
	konnectDiffCmd.Flags().IntVar(&konnectDiffCmdParallelism, "parallelism",
		100, "Maximum number of concurrent operations")
	konnectDiffCmd.Flags().BoolVar(&konnectDiffCmdNonZeroExitCode, "non-zero-exit-code",
		false, "return exit code 2 if there is a diff present,\n"+
			"exit code 0 if no diff is found,\n"+
			"and exit code 1 if an error occurs.")
	addSilenceEventsFlag(konnectDiffCmd.Flags())
}
