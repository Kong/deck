package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newDebriefCmd represents the debrief command
func newDebriefCmd() *cobra.Command {
	var debriefCmdKongStateFile []string
	var debriefCmdLong bool
	debriefCmd := &cobra.Command{
		Use:   "debrief",
		Short: "Debrief provides a quick summary from one or more decK files",
		Long: `The debrief command reads one or more state files, and provides an overall summary of
the services and plugins used.`,
		Args: validateNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return debriefMain(cmd.Context(), debriefCmdKongStateFile, debriefCmdLong)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(debriefCmdKongStateFile) == 0 {
				return fmt.Errorf("one or more state files with Kong's configuration " +
					"must be specified using `-s`/`--state` flag")
			}
			return preRunSilenceEventsFlag()
		},
	}

	debriefCmd.Flags().StringSliceVarP(&debriefCmdKongStateFile,
		"state", "s", []string{"kong.yaml"}, "file(s) containing Kong's configuration.\n"+
			"This flag can be specified multiple times for multiple files.\n"+
			"Use `-` to read from stdin.")
	debriefCmd.Flags().BoolVarP(&debriefCmdLong,
		"long", "l", false, "long debrief, shows additional details")

	addSilenceEventsFlag(debriefCmd.Flags())
	return debriefCmd
}
