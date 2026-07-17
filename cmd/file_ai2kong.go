package cmd

import (
	"fmt"
	"os"
	"strings"

	ai2kong "github.com/Kong/ai-deck-converter/convert"
	"github.com/kong/go-apiops/filebasics"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var (
	convertSourceFile   string
	convertOutputFile   string
	convertOutputFormat string
)

const managedByAIDeckTag = "managed_by:deck-ai"

func newAi2KongCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai2kong",
		Short: "Generate Kong configuration from AI Gateway configuration",
		Long: `This command takes an AI Gateway 2.0 entity model and converts it to a standard decK state file.` +
			` It also adds the 'managed_by:deck-ai' tag which is used internally to all entities by default.` +
			"\n\nThe source file may be provided in either YAML or JSON; the format is auto-detected." +
			" The output format is controlled by the --format flag.",
		Args:    validateNoArgs,
		PreRunE: validateAi2KongFlags,
		RunE:    execute,
	}

	cmd.Flags().StringVarP(&convertSourceFile, "source", "s", "", "AI Gateway source file (required)")
	cmd.Flags().StringVarP(&convertOutputFile, "output-file", "o", "",
		"Output Kong decK file (optional, defaults to stdout)")
	cmd.Flags().StringVarP(&convertOutputFormat, "format", "", "yaml", "output format: yaml or json")

	return cmd
}

func validateAi2KongFlags(_ *cobra.Command, _ []string) error {
	if convertSourceFile == "" {
		return fmt.Errorf("--source/-s flag is required")
	}
	return nil
}

func execute(cmd *cobra.Command, _ []string) error {
	_ = sendAnalytics("file-ai2kong", "", modeAIGateway)

	format := strings.ToLower(getFormatFlagValue(cmd, convertOutputFormat))

	// Read source file (auto-detects YAML or JSON).
	sourceContent, err := filebasics.ReadFile(convertSourceFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	converted, warnings, err := ai2kong.Convert(sourceContent, ai2kong.Options{
		OutputMode: "deck",
	})
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	printAIWarnings(os.Stderr, warnings)

	// Add the default select_tags to the converted document's _info section
	doc, err := addDefaultSelectTags(converted)
	if err != nil {
		return err
	}

	// Write output in the requested format. An empty output file means stdout,
	// which filebasics represents as "-".
	outputFile := convertOutputFile
	if outputFile == "" {
		outputFile = "-"
	}
	if err := filebasics.WriteSerializedFile(outputFile, doc, filebasics.OutputFormat(format)); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// addDefaultSelectTags parses the converted YAML and ensures the _info section
// carries the default select_tags, creating _info if it is absent. It returns
// the mutated document as a map ready for serialization.
func addDefaultSelectTags(converted []byte) (map[string]interface{}, error) {
	var docMap map[string]interface{}
	if err := yaml.Unmarshal(converted, &docMap); err != nil {
		return nil, fmt.Errorf("failed to parse converted config: %w", err)
	}

	if infoMap, ok := docMap["_info"].(map[string]interface{}); ok {
		// _info exists, update select_tags
		infoMap["select_tags"] = []string{managedByAIDeckTag}
	} else {
		// _info doesn't exist, create it with select_tags
		docMap["_info"] = map[string]interface{}{
			"select_tags": []string{managedByAIDeckTag},
		}
	}

	return docMap, nil
}
