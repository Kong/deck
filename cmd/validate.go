package cmd

import (
	"fmt"

	"github.com/kong/deck/dump"
	"github.com/kong/deck/file"
	"github.com/kong/deck/state"
	"github.com/spf13/cobra"
)

var (
	validateCmdKongStateFile     []string
	validateCmdRBACResourcesOnly bool
)

// validateCmd represents the diff command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the state file",
	Long: `The validate command reads the state file and ensures validity.

It reads all the specified state files and reports YAML/JSON
parsing issues. It also checks for foreign relationships
and alerts if there are broken relationships, or missing links present.
No communication takes places between decK and Kong during the execution of
this command.
`,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = sendAnalytics("validate", "")
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
		}, dump.Config{})
		if err != nil {
			return err
		}
		if err := checkForRBACResources(*rawState, validateCmdRBACResourcesOnly); err != nil {
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
			return fmt.Errorf("a state file with Kong's configuration " +
				"must be specified using -s/--state flag")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVar(&validateCmdRBACResourcesOnly, "rbac-resources-only",
		false, "indicate that the state file(s) contains RBAC resources only (Kong Enterprise only).")
	validateCmd.Flags().StringSliceVarP(&validateCmdKongStateFile,
		"state", "s", []string{"kong.yaml"}, "file(s) containing Kong's configuration.\n"+
			"This flag can be specified multiple times for multiple files.\n"+
			"Use '-' to read from stdin.")
}
