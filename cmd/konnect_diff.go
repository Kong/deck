package cmd

import (
	"github.com/spf13/cobra"
)

var (
	konnectDiffCmdKongStateFile   []string
	konnectDiffCmdParallelism     int
	konnectDiffCmdNonZeroExitCode bool
)

// newKonnectDiffCmd represents the 'deck konnect diff' command.
func newKonnectDiffCmd() *cobra.Command {
	konnectDiffCmd := &cobra.Command{
		Use:   "diff",
		Short: "Diff the current entities in Konnect with the one on disks (in alpha)",
		Long: `The konnect diff command is similar to a dry run of the 'deck konnect sync' command.

	It loads entities from Konnect and performs a diff with
	the entities in local files. This allows you to see the entities
	that will be created, updated, or deleted.` + konnectAlphaState,
		Args: validateNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = sendAnalytics("konnect-diff", "", modeKonnect)
			return syncKonnect(cmd.Context(), konnectDiffCmdKongStateFile, true,
				konnectDiffCmdParallelism)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRunSilenceEventsFlag()
		},
	}

	konnectDiffCmd.Flags().StringSliceVarP(&konnectDiffCmdKongStateFile,
		"state", "s", []string{"konnect.yaml"}, "file(s) containing Konnect's configuration.\n"+
			"This flag can be specified multiple times for multiple files.")
	konnectDiffCmd.Flags().BoolVar(&konnectDumpIncludeConsumers, "include-consumers",
		false, "export consumers, associated credentials and any plugins associated "+
			"with consumers.")
	konnectDiffCmd.Flags().IntVar(&konnectDiffCmdParallelism, "parallelism",
		100, "Maximum number of concurrent operations.")
	konnectDiffCmd.Flags().BoolVar(&noMaskValues, "no-mask-deck-env-vars-value",
		false, "do not mask DECK_ environment variable values at diff output.")
	konnectDiffCmd.Flags().BoolVar(&konnectDiffCmdNonZeroExitCode, "non-zero-exit-code",
		false, "return exit code 2 if there is a diff present,\n"+
			"exit code 0 if no diff is found,\n"+
			"and exit code 1 if an error occurs.")
	addSilenceEventsFlag(konnectDiffCmd.Flags())
	return konnectDiffCmd
}
