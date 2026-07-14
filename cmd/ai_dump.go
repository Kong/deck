package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/Kong/ai-deck-converter/revert"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/state"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var (
	aiDumpCmdStateFormat = "yaml" // default to yaml
	aiDumpCmdOutputFile  string
	aiDumpWorkspace      string
)

func executeAiDump(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	if yes, err := utils.ConfirmFileOverwrite(aiDumpCmdOutputFile, aiDumpCmdStateFormat, assumeYes); err != nil {
		return err
	} else if !yes {
		return nil
	}

	// Set the selector tags to only get AI-managed entities
	originalTags := dumpConfig.SelectorTags
	dumpConfig.SelectorTags = []string{aiManagedSelectorTag}
	defer func() {
		dumpConfig.SelectorTags = originalTags
	}()

	wsClient, err := utils.GetKongClient(rootConfig)
	if err != nil {
		return err
	}

	kongVersion, err := fetchKongVersion(ctx, rootConfig.ForWorkspace(aiDumpWorkspace))
	if err != nil {
		return fmt.Errorf("reading Kong version: %w", err)
	}

	isAIGateway, err := kong.IsKongAIGateway()
	if err != nil {
		return fmt.Errorf("checking if Kong is an AI Gateway: %w", err)
	}

	writeConfig := file.WriteConfig{
		SelectTags:                       dumpConfig.SelectorTags,
		Workspace:                        aiDumpWorkspace,
		Filename:                         "",
		WithID:                           false,
		KongVersion:                      kongVersion,
		IsKongAIGateway:                  isAIGateway,
		IsConsumerGroupPolicyOverrideSet: dumpConfig.IsConsumerGroupPolicyOverrideSet,
		SanitizeContent:                  false,
		IncludePluginDefinitions:         dumpConfig.IncludePluginDefinitions,
	}

	if aiDumpWorkspace != "" {
		wsClient, err = getWorkspaceClient(ctx, aiDumpWorkspace)
		if err != nil {
			return err
		}
	}

	// Get Kong state with the AI tag selector
	ks, err := getKongState(ctx, wsClient)
	if err != nil {
		return fmt.Errorf("getting Kong state: %w", err)
	}

	// Convert Kong state to YAML in memory
	kongYAML, err := kongsStateToYAML(ks, writeConfig)
	if err != nil {
		return fmt.Errorf("converting Kong state to YAML: %w", err)
	}

	// Revert Kong YAML back to AI Gateway format
	aiGatewayYAML, warnings, err := revert.Revert(kongYAML, revert.Options{})
	if err != nil {
		return fmt.Errorf("reverting Kong configuration to AI Gateway format: %w", err)
	}

	// Print warnings to stderr
	if len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", warning)
		}
	}

	// Parse the AI Gateway YAML to add select_tags to _info
	var doc interface{}
	err = yaml.Unmarshal(aiGatewayYAML, &doc)
	if err != nil {
		return fmt.Errorf("failed to parse AI Gateway YAML: %w", err)
	}

	// Ensure the document is a map and add select_tags to _info
	if docMap, ok := doc.(map[string]interface{}); ok {
		if infoMap, ok := docMap["_info"].(map[string]interface{}); ok {
			// _info exists, update select_tags
			infoMap["select_tags"] = []string{aiManagedSelectorTag}
		} else {
			// _info doesn't exist, create it with select_tags
			docMap["_info"] = map[string]interface{}{
				"select_tags": []string{aiManagedSelectorTag},
			}
		}
	}

	// Marshal back to YAML
	finalOutput, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal modified AI Gateway document: %w", err)
	}

	// Write output
	var output io.Writer
	if aiDumpCmdOutputFile != "" && aiDumpCmdOutputFile != "-" {
		outFile, err := os.Create(aiDumpCmdOutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outFile.Close()
		output = outFile
	} else {
		output = os.Stdout
	}

	_, err = output.Write(finalOutput)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// kongsStateToYAML converts Kong state to YAML bytes
func kongsStateToYAML(ks *state.KongState, cfg file.WriteConfig) ([]byte, error) {
	fileContent, err := file.KongStateToContent(ks, cfg)
	if err != nil {
		return nil, fmt.Errorf("converting Kong state to file content: %w", err)
	}

	return yaml.Marshal(fileContent)
}

func newAiDumpCmd() *cobra.Command {
	aiDumpCmd := &cobra.Command{
		Use:   "dump",
		Short: "Dump AI Gateway configuration from Kong (managed by deck-ai)",
		Long: `The ai dump command reads AI Gateway entities (tagged with 'managed-by: deck-ai')
from Kong and writes them to a local file in AI Gateway format.

The command exports only entities that have the 'managed-by: deck-ai' tag.`,
		Args: validateNoArgs,
		RunE: executeAiDump,
	}

	aiDumpCmd.Flags().StringVarP(&aiDumpWorkspace, "workspace", "w",
		"", "dump configuration of a specific Workspace (Kong Enterprise only).")
	aiDumpCmd.Flags().StringVarP(&aiDumpCmdOutputFile, "output-file", "o",
		"-", "file to which to write AI Gateway configuration. Use `-` to write to stdout.")
	aiDumpCmd.Flags().BoolVar(&assumeYes, "yes",
		false, "assume `yes` to prompts and run non-interactively.")

	return aiDumpCmd
}
