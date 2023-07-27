package cmd

import (
	"strings"

	"github.com/kong/deck/convert"
	"github.com/kong/deck/file"
	"github.com/spf13/cobra"
)

var (
	fileRenderCmdKongStateFile  []string
	fileRenderCmdKongFileOutput string
	fileRenderCmdStateFormat    string
)

func executeFileRenderCmd(_ *cobra.Command, _ []string) error {
	return convert.Convert(
		fileRenderCmdKongStateFile,
		fileRenderCmdKongFileOutput,
		file.Format(strings.ToUpper(fileRenderCmdStateFormat)),
		convert.FormatDistributed,
		convert.FormatKongGateway3x,
		true)
}

func newFileRenderCmd() *cobra.Command {
	renderCmd := &cobra.Command{
		Use:   "render",
		Short: "Render the configuration as Kong declarative config",
		Long:  ``,
		Args:  cobra.ArbitraryArgs,
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
	renderCmd.Flags().StringVar(&fileRenderCmdStateFormat, "format",
		"yaml", "output file format: json or yaml.")

	return renderCmd
}
