package cmd

import (
	"github.com/kong/deck/convert"
	"github.com/spf13/cobra"
)

var (
	fileRenderCmdKongStateFile  []string
	fileRenderCmdKongFileOutput string
)

func executeFileRenderCmd(_ *cobra.Command, _ []string) error {
	return convert.Convert(fileRenderCmdKongStateFile, fileRenderCmdKongFileOutput,
		convert.FormatDistributed, convert.FormatKongGateway3x, true)
}

func newFileRenderCmd() *cobra.Command {
	renderCmd := &cobra.Command{
		Use:   "render",
		Short: "Render the configuration as Kong declarative config",
		Long:  ``,
		Args:  validateNoArgs,
		RunE:  executeFileRenderCmd,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			fileRenderCmdKongStateFile = args
			if len(fileRenderCmdKongStateFile) == 0 {
				fileRenderCmdKongStateFile = []string{"-"} // default to stdin
			}
			return preRunSilenceEventsFlag()
		},
	}

	renderCmd.Flags().StringVarP(&fileRenderCmdKongFileOutput, "output-file", "o",
		"-", "file to which to write Kong's configuration."+
			"Use `-` to write to stdout.")

	// TODO: support json output
	// renderCmd.Flags().StringVar(&fileRenderCmdStateFormat, "format",
	// 	"yaml", "output file format: json or yaml.")

	return renderCmd
}
