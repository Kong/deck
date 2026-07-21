package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	ai2kong "github.com/Kong/ai-deck-converter/revert"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/state"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var (
	aiDumpCmdStateFormat string
	aiDumpCmdOutputFile  string
	aiDumpWorkspace      string
)

func executeAiDump(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	format := strings.ToLower(getFormatFlagValue(cmd, aiDumpCmdStateFormat))

	if yes, err := utils.ConfirmFileOverwrite(aiDumpCmdOutputFile, format, assumeYes); err != nil {
		return err
	} else if !yes {
		return nil
	}

	// AI Gateway entities (tagged managed_by:deck-ai) must be managed with
	// kongctl on Konnect, not decK. Since ai dump only ever reads AI-managed
	// entities, fail fast before any Konnect calls.
	if inKonnectMode(nil) {
		return errAIManagedEntitiesOnKonnect()
	}

	// Set the selector tags to only get AI-managed entities
	dumpConfig.SelectorTags = []string{managedByAIDeckTag}

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
		IsConsumerGroupPolicyOverrideSet: false,
		SanitizeContent:                  false,
		IncludePluginDefinitions:         false,
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
	aiGatewayYAML, warnings, err := ai2kong.Revert(kongYAML, ai2kong.Options{})
	if err != nil {
		return fmt.Errorf("reverting Kong configuration to AI Gateway format: %w", err)
	}

	printAIWarnings(os.Stderr, warnings)

	outputBytes, err := aiDumpOutput(aiGatewayYAML, format)
	if err != nil {
		return err
	}

	// Write output
	var output io.Writer = os.Stdout
	if aiDumpCmdOutputFile != "" && aiDumpCmdOutputFile != "-" {
		outFile, err := os.Create(aiDumpCmdOutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outFile.Close()
		output = outFile
	}

	_, err = output.Write(outputBytes)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// aiDumpOutput returns the AI Gateway configuration serialized in the requested
// format, which is expected to be lower-cased by the caller. revert.Revert
// always produces YAML, so YAML is returned verbatim (to preserve the library's
// formatting) while JSON is produced by re-serializing.
func aiDumpOutput(aiGatewayYAML []byte, format string) ([]byte, error) {
	switch filebasics.OutputFormat(format) {
	case filebasics.OutputFormatYaml:
		return aiGatewayYAML, nil
	case filebasics.OutputFormatJSON:
		m, err := filebasics.Deserialize(aiGatewayYAML)
		if err != nil {
			return nil, fmt.Errorf("parsing reverted configuration: %w", err)
		}
		out, err := filebasics.Serialize(m, filebasics.OutputFormatJSON)
		if err != nil {
			return nil, fmt.Errorf("serializing configuration to JSON: %w", err)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("expected format to be either %q or %q, got: %q",
			filebasics.OutputFormatYaml, filebasics.OutputFormatJSON, format)
	}
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
		Long: `The ai dump command reads AI Gateway entities (tagged with 'managed_by:deck-ai')
from Kong and writes them to a local file in AI Gateway format.

The command exports only entities that have the 'managed_by:deck-ai' tag.

The output can be written as either YAML or JSON, controlled by the --format flag.`,
		Args: validateNoArgs,
		RunE: executeAiDump,
	}

	aiDumpCmd.Flags().StringVarP(&aiDumpWorkspace, "workspace", "w",
		"", "dump configuration of a specific Workspace (Kong Enterprise only).")
	aiDumpCmd.Flags().StringVarP(&aiDumpCmdOutputFile, "output-file", "o",
		"-", "file to which to write AI Gateway configuration. Use `-` to write to stdout.")
	aiDumpCmd.Flags().StringVar(&aiDumpCmdStateFormat, "format",
		string(filebasics.OutputFormatYaml), "output file format: json or yaml.")
	aiDumpCmd.Flags().BoolVar(&assumeYes, "yes",
		false, "assume `yes` to prompts and run non-interactively.")

	return aiDumpCmd
}
