package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/Kong/ai-deck-converter/convert"
	"github.com/spf13/cobra"
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

	// Convert AI Gateway to Kong decK
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

	// Write output
	var output io.Writer
	if convertOutputFile != "" {
		outFile, err := os.Create(convertOutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outFile.Close()
		output = outFile
	} else {
		output = os.Stdout
	}

	_, err = output.Write(converted)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
