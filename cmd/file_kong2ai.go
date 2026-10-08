package cmd

import (
	"fmt"
	"os"
	"strings"

	kong2ai "github.com/Kong/ai-deck-converter/revert"
	"github.com/kong/go-apiops/filebasics"
	"github.com/spf13/cobra"
)

var (
	kong2AiSourceFile   string
	kong2AiOutputFile   string
	kong2AiOutputFormat string
)

func newKong2AiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kong2ai",
		Short: "Generate AI Gateway configuration from Kong configuration",
		Long: `This command takes a standard decK state file and converts it to an AI Gateway 2.0 entity model.` +
			` It helps migrate an existing Kong Gateway configuration to AI Gateway,` +
			` and is the inverse of 'deck file ai2kong'.` +
			"\n\nThe conversion is best-effort: AI plugins are recognized anywhere in the configuration," +
			" while entities with no AI Gateway representation are reported as warnings." +
			"\n\nThe source file may be provided in either YAML or JSON; the format is auto-detected." +
			" The output format is controlled by the --format flag.",
		Args:    validateNoArgs,
		PreRunE: validateKong2AiFlags,
		RunE:    executeFileKong2Ai,
	}

	cmd.Flags().StringVarP(&kong2AiSourceFile, "source", "s", "", "Kong decK source file (required)")
	cmd.Flags().StringVarP(&kong2AiOutputFile, "output-file", "o", "",
		"Output AI Gateway file (optional, defaults to stdout)")
	cmd.Flags().StringVar(&kong2AiOutputFormat, "format",
		string(filebasics.OutputFormatYaml), "output format: yaml or json")

	return cmd
}

func validateKong2AiFlags(_ *cobra.Command, _ []string) error {
	if kong2AiSourceFile == "" {
		return fmt.Errorf("--source/-s flag is required")
	}
	return nil
}

func executeFileKong2Ai(cmd *cobra.Command, _ []string) error {
	_ = sendAnalytics("file-kong2ai", "", modeAIGateway)

	format := strings.ToLower(getFormatFlagValue(cmd, kong2AiOutputFormat))

	sourceContent, err := filebasics.ReadFile(kong2AiSourceFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	aiGatewayYAML, warnings, err := kong2ai.Revert(sourceContent, kong2ai.Options{})
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	printAIWarnings(os.Stderr, warnings)

	outputBytes, err := aiDumpOutput(aiGatewayYAML, format)
	if err != nil {
		return err
	}

	// An empty output file means stdout, which filebasics represents as "-".
	outputFile := kong2AiOutputFile
	if outputFile == "" {
		outputFile = "-"
	}
	if err := filebasics.WriteFile(outputFile, outputBytes); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
