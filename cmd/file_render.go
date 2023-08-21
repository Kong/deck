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
		Short: "Combines multiple complete configuration files into one Kong declarative config file.",
		Long:  `Combines multiple complete configuration files into one Kong
	declarative config file.

 This command can render the output in JSON or YAML format. Unlike 
 "deck file merge", the render command accepts complete configuration files, 
 while "deck file merge" only combines partial file snippets.

 For example, the following command takes two input files and renders them as one 
 combined JSON file:

   deck file render kong1.yml kong2.yml -o kong3 --format json
	`,
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
