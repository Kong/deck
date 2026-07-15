package cmd

import (
	"fmt"
	"io"
	"os"

	ai2kong "github.com/Kong/ai-deck-converter/convert"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var (
	convertSourceFile  string
	convertOutputFile  string
	managedByAIDeckTag = "managed_by:deck-ai"
)

func newAi2KongCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai2kong",
		Short: "Generate Kong configuration from AI Gateway configuration",
		Long: `This command takes an AI Gateway 2.0 entity model and converts it to a standard decK state file.` +
			` It also adds the 'managed_by:deck-ai' tag which is used internally to all entities by default.`,
		Args:    validateNoArgs,
		PreRunE: validateAi2KongFlags,
		RunE:    execute,
	}

	cmd.Flags().StringVarP(&convertSourceFile, "source", "s", "", "AI Gateway source file (required)")
	cmd.Flags().StringVarP(&convertOutputFile, "output-file", "o", "",
		"Output Kong decK YAML file (optional, defaults to stdout)")

	return cmd
}

func validateAi2KongFlags(_ *cobra.Command, _ []string) error {
	if convertSourceFile == "" {
		return fmt.Errorf("--source/-s flag is required")
	}
	return nil
}

func execute(_ *cobra.Command, _ []string) error {
	// Read source file
	sourceContent, err := os.ReadFile(convertSourceFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	converted, warnings, err := ai2kong.Convert(sourceContent, ai2kong.Options{
		OutputMode: "deck",
	})
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Print warnings to stderr
	if len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", warning)
		}
	}

	// Add the default select_tags to the converted document's _info section
	output, err := addDefaultSelectTags(converted)
	if err != nil {
		return err
	}

	// Write output
	var outputWriter io.Writer
	outputWriter = os.Stdout

	if convertOutputFile != "" {
		outFile, err := os.Create(convertOutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outFile.Close()
		outputWriter = outFile
	}

	_, err = outputWriter.Write(output)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// addDefaultSelectTags parses the converted YAML and ensures the _info section
// carries the default select_tags, creating _info if it is absent. It returns
// the re-marshaled YAML.
func addDefaultSelectTags(converted []byte) ([]byte, error) {
	var doc interface{}
	if err := yaml.Unmarshal(converted, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse converted YAML: %w", err)
	}

	// Ensure the document is a map and add select_tags to _info
	if docMap, ok := doc.(map[string]interface{}); ok {
		if infoMap, ok := docMap["_info"].(map[string]interface{}); ok {
			// _info exists, update select_tags
			infoMap["select_tags"] = []string{managedByAIDeckTag}
		} else {
			// _info doesn't exist, create it with select_tags
			docMap["_info"] = map[string]interface{}{
				"select_tags": []string{managedByAIDeckTag},
			}
		}
	}

	output, err := marshalToYAML(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modified document: %w", err)
	}
	return output, nil
}

// marshalToYAML encodes v as YAML using a two-space indent.
func marshalToYAML(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}
