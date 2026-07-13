package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/Kong/ai-deck-converter/convert"
	"github.com/kong/go-database-reconciler/pkg/cprint"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/spf13/cobra"
)

var (
	aiSyncSourceFile    string
	aiSyncWorkspace     string
	aiSyncParallelism   int
	aiSyncDBUpdateDelay int
	aiSyncJSONOutput    bool
)

func executeAiSync(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	if aiSyncJSONOutput {
		initJSONOutput()
	}

	sourceContent, err := readAiSyncSource(aiSyncSourceFile)
	if err != nil {
		return err
	}

	convertedYAML, warnings, err := convert.Convert(sourceContent, convert.Options{})
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}
	reportAiConversionWarnings(warnings, aiSyncJSONOutput)

	targetContent, err := file.GetContentFromReader(bytes.NewReader(convertedYAML), file.EnvVarsSkip)
	if err != nil {
		return fmt.Errorf("parsing converted configuration: %w", err)
	}

	injectAiManagedSelectorTag(targetContent)

	return syncContent(ctx, targetContent, false, aiSyncParallelism, aiSyncDBUpdateDelay,
		aiSyncWorkspace, aiSyncJSONOutput, ApplyTypeFull)
}

// readAiSyncSource reads the AI Gateway source configuration from filename,
// or from stdin when filename is "-".
func readAiSyncSource(filename string) ([]byte, error) {
	if filename == "-" {
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}
		return content, nil
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read source file: %w", err)
	}
	return content, nil
}

// injectAiManagedSelectorTag ensures the AI-managed select tag is present in
// _info.select_tags, adding it without discarding any tags already declared.
// This scopes the sync (including pruning) to AI Gateway entities only, the same
// way `ai dump` scopes its read.
func injectAiManagedSelectorTag(content *file.Content) {
	if content.Info == nil {
		content.Info = &file.Info{}
	}
	if slices.Contains(content.Info.SelectorTags, aiManagedSelectorTag) {
		return
	}
	content.Info.SelectorTags = append(content.Info.SelectorTags, aiManagedSelectorTag)
}

func newAiSyncCmd() *cobra.Command {
	aiSyncCmd := &cobra.Command{
		Use:   "sync [flags] [ai-gateway-state-file]",
		Short: "Sync AI Gateway configuration to Kong",
		Long: `The ai sync command converts an AI Gateway configuration file to Kong
configuration and syncs it directly to Kong, tagging every managed entity with
'managed-by: deck-ai'.

The AI Gateway state file is provided as a positional argument. Use '-' to read
from stdin (the default when no argument is given).

This is the direct equivalent of running 'deck file ai2kong' followed by
'deck gateway sync' on the result.`,
		Args: cobra.MaximumNArgs(1),
		PreRunE: func(_ *cobra.Command, args []string) error {
			aiSyncSourceFile = "-"
			if len(args) > 0 {
				aiSyncSourceFile = args[0]
			}
			return checkParallelism(aiSyncParallelism)
		},
		RunE: executeAiSync,
	}

	aiSyncCmd.Flags().StringVarP(&aiSyncWorkspace, "workspace", "w", "",
		"Sync configuration to a specific workspace "+
			"(Kong Enterprise only).\n"+
			"This takes precedence over _workspace fields in state files.")
	aiSyncCmd.Flags().IntVar(&aiSyncParallelism, "parallelism",
		10, "Maximum number of concurrent operations.")
	aiSyncCmd.Flags().IntVar(&aiSyncDBUpdateDelay, "db-update-propagation-delay",
		0, "artificial delay (in seconds) that is injected between insert operations \n"+
			"for related entities (usually for Cassandra deployments).\n"+
			"See `db_update_propagation` in kong.conf.")
	aiSyncCmd.Flags().BoolVar(&syncCmdAssumeYes, "yes",
		false, "assume `yes` to prompts and run non-interactively.")
	aiSyncCmd.Flags().BoolVar(&aiSyncJSONOutput, "json-output",
		false, "generate command execution report in a JSON format.")

	return aiSyncCmd
}

// reportAiConversionWarnings surfaces warnings produced while converting an
// AI Gateway file. When JSON output is enabled they are collected into the
// structured report; otherwise they are printed to stderr.
func reportAiConversionWarnings(warnings []string, jsonOutputEnabled bool) {
	for _, warning := range warnings {
		if jsonOutputEnabled {
			jsonOutput.Warnings = append(jsonOutput.Warnings, warning)
		} else {
			cprint.DeletePrintf("Warning: " + warning + "\n")
		}
	}
}
