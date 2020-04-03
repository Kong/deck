package cmd

import (
	"github.com/hbagdi/deck/file"
	"github.com/hbagdi/deck/state"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	validateCmdKongStateFile []string
)

// validateCmd represents the diff command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the state file",
	Long: `Validate reads the state file and ensures the validity.

It will read all the state files that are passed in. If there are YAML/JSON
parsing issues, they will be reported. It also checks for foreign relationships
and alerts if there are broken relationships, missing links present.
No communication takes places between decK and Kong during the execution of
this command.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// read target file
		// this does json schema validation as well
		targetContent, err := file.GetContentFromFiles(validateCmdKongStateFile)
		if err != nil {
			return err
		}

		dummyEmptyState, err := state.NewKongState()
		if err != nil {
			return err
		}

		rawState, err := file.Get(targetContent, file.RenderConfig{
			CurrentState: dummyEmptyState,
		})
		if err != nil {
			return err
		}
		// this catches foreign relation errors
		_, err = state.Get(rawState)
		if err != nil {
			return err
		}

		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(validateCmdKongStateFile) == 0 {
			return errors.New("A state file with Kong's configuration " +
				"must be specified using -s/--state flag.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringSliceVarP(&validateCmdKongStateFile,
		"state", "s", []string{"kong.yaml"}, "file(s) containing Kong's configuration.\n"+
			"This flag can be specified multiple times for multiple files.\n"+
			"Use '-' to read from stdin.")
}
