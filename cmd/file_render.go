package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/kong/deck/file"
	"github.com/kong/deck/state"
	"github.com/spf13/cobra"
)

var (
	fileRenderCmdKongStateFile  []string
	fileRenderCmdKongFileOutput string
	fileRenderCmdStateFormat    string
)

func newFileRenderCmd() *cobra.Command {
	renderCmd := &cobra.Command{
		Use:   "render",
		Short: "Render the configuration as Kong declarative config",
		Long:  ``,
		Args:  validateNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return render(cmd.Context(), fileRenderCmdKongStateFile, fileRenderCmdStateFormat)
		},
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

func render(ctx context.Context, filenames []string, format string) error {
	targetContent, err := file.GetContentFromFiles(filenames, true)
	if err != nil {
		return err
	}
	s, _ := state.NewKongState()
	rawState, err := file.Get(ctx, targetContent, file.RenderConfig{
		CurrentState: s,
		KongVersion:  semver.Version{Major: 3, Minor: 0},
	}, dumpConfig, nil)
	if err != nil {
		return err
	}
	targetState, err := state.Get(rawState)
	if err != nil {
		return err
	}

	return file.KongStateToFile(targetState, file.WriteConfig{
		Filename:    fileRenderCmdKongFileOutput,
		FileFormat:  file.Format(strings.ToUpper(format)),
		KongVersion: "3.0.0",
	})
}
