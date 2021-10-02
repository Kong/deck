package cmd

import (
	"context"
	"fmt"
	"sync"

	"github.com/fatih/color"
	"github.com/kong/deck/file"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/spf13/cobra"
)

var (
	validateCmdKongStateFile     []string
	validateCmdRBACResourcesOnly bool
	validateCmdKongAPI           bool
	validateCmdKongAPIEntity     []string
)

// validateCmd represents the diff command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long: `The validate command supports 2 modes to ensure validity:
	- reads the state file (default)
	- queries Kong API

By default, it reads all the specified state files and reports YAML/JSON
parsing issues. It also checks for foreign relationships
and alerts if there are broken relationships, or missing links present.

When ran with the --use-kong-api flag, it uses the 'validate' endpoint in Kong
to validate entities configuration.
`,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = sendAnalytics("validate", "")
		if validateCmdKongAPI {
			return validateWithKong(cmd, validateCmdKongAPIEntity)
		}
		return validateWithFile()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(validateCmdKongStateFile) == 0 {
			return fmt.Errorf("a state file with Kong's configuration " +
				"must be specified using -s/--state flag")
		} else if validateCmdKongAPI && len(validateCmdKongAPIEntity) == 0 {
			return fmt.Errorf("a list of entities must be specified " +
				"using the -e/--entity flag when validating via Kong API " +
				"(e.g. -e entity1 -e entity2)")
		}
		return nil
	},
}

func validate(ctx context.Context, entity string, kongClient *kong.Client) string {
	_, err := performValidate(ctx, kongClient, entity)
	output := color.New(color.FgGreen, color.Bold).Sprintf("%s: schema validation successful", entity)
	if err != nil {
		output = color.New(color.FgRed, color.Bold).Sprintf("%s: %v", entity, err)
	}
	return output
}

func validateWithKong(cmd *cobra.Command, entity []string) error {
	ctx := cmd.Context()
	kongClient, err := utils.GetKongClient(rootConfig)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(len(entity))

	for _, e := range entity {
		go func(e string) {
			defer wg.Done()
			output := validate(ctx, e, kongClient)
			fmt.Println(output)
		}(e)
	}
	wg.Wait()

	return nil
}

func validateWithFile() error {
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
	if err := checkForRBACResources(*rawState, validateCmdRBACResourcesOnly); err != nil {
		return err
	}
	// this catches foreign relation errors
	_, err = state.Get(rawState)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVar(&validateCmdRBACResourcesOnly, "rbac-resources-only",
		false, "indicate that the state file(s) contains RBAC resources only (Kong Enterprise only).")
	validateCmd.Flags().StringSliceVarP(&validateCmdKongStateFile,
		"state", "s", []string{"kong.yaml"}, "file(s) containing Kong's configuration.\n"+
			"This flag can be specified multiple times for multiple files.\n"+
			"Use '-' to read from stdin.")
	validateCmd.Flags().BoolVar(&validateCmdKongAPI, "use-kong-api",
		false, "whether to leverage Kong's API to validate configuration.")
	validateCmd.Flags().StringSliceVarP(&validateCmdKongAPIEntity,
		"entity", "e", []string{}, "entities to be validated via Kong API "+
			"(e.g. -e entity1 -e entity2)")
}
