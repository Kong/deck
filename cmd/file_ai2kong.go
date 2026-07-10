package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/Kong/ai-deck-converter/convert"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var (
	convertSourceFile string
	convertOutputFile string
)

func newAi2KongCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ai2kong",
		Short:   "Generate Kong configuration from AI Gateway configuration",
		Long:    `This command takes an AI Gateway 2.0 entity model and converts it to a standard decK state file`,
		Args:    validateNoArgs,
		PreRunE: validateAi2KongFlags,
		RunE:    execute,
	}

	cmd.Flags().StringVarP(&convertSourceFile, "state", "s", "", "AI Gateway state file (required)")
	cmd.Flags().StringVarP(&convertOutputFile, "output-file", "o", "", "Output Kong decK YAML file (optional, defaults to stdout)")

	return cmd
}

func validateAi2KongFlags(cmd *cobra.Command, args []string) error {
	if convertSourceFile == "" {
		return fmt.Errorf("--state/-s flag is required")
	}
	return nil
}

func execute(cmd *cobra.Command, args []string) error {
	// Read source file
	sourceContent, err := os.ReadFile(convertSourceFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Convert AI Gateway to Kong decK (without GlobalSelectTags option)
	converted, warnings, err := convert.Convert(sourceContent, convert.Options{})
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Print warnings to stderr
	if len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", warning)
		}
	}

	// Parse the converted YAML to add select_tags to _info
	var doc interface{}
	err = yaml.Unmarshal(converted, &doc)
	if err != nil {
		return fmt.Errorf("failed to parse converted YAML: %w", err)
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
	output, err := marshalToYAML(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal modified document: %w", err)
	}

	// Write output
	var outputWriter io.Writer
	if convertOutputFile != "" {
		outFile, err := os.Create(convertOutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outFile.Close()
		outputWriter = outFile
	} else {
		outputWriter = os.Stdout
	}

	_, err = outputWriter.Write(output)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// marshalToYAML encodes v as YAML using a two-space indent.
func marshalToYAML(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}
