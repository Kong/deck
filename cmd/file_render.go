package cmd

import (
	"fmt"

	"github.com/kong/deck/convert"
	"github.com/spf13/cobra"
)

var (
	fileRenderCmdKongStateFile  []string
	fileRenderCmdKongFileOutput string
	fileRenderCmdStateFormat    string
)

func executeFileRenderCmd(_ *cobra.Command, _ []string) error {
	destinationFormat, err := convert.ParseFormat(fileRenderCmdStateFormat)
	if err != nil {
		return err
	}

	return convert.Convert(fileRenderCmdKongStateFile, fileRenderCmdKongFileOutput,
		convert.FormatDistributed, destinationFormat, true)
}

func newFileRenderCmd() *cobra.Command {
	renderCmd := &cobra.Command{
		Use:   "render",
		Short: "Render the configuration as Kong declarative config",
		Long:  ``,
		Args:  validateNoArgs,
		RunE:  executeFileRenderCmd,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(fileRenderCmdStateFormat) == 0 {
				return fmt.Errorf("a state file with Kong's configuration " +
					"must be specified using `-s`/`--state` flag")
			}
			return preRunSilenceEventsFlag()
		},
	}

	renderCmd.Flags().StringSliceVarP(&fileRenderCmdKongStateFile,
		"state", "s", []string{"-"}, "file(s) containing Kong's configuration.\n"+
			"This flag can be specified multiple times for multiple files.\n"+
			"Use `-` to read from stdin.")
	renderCmd.Flags().StringVarP(&fileRenderCmdKongFileOutput, "output-file", "o",
		"-", "file to which to write Kong's configuration."+
			"Use `-` to write to stdout.")
	renderCmd.Flags().StringVar(&fileRenderCmdStateFormat, "format",
		"yaml", "output file format: json or yaml.")

	return renderCmd
}
