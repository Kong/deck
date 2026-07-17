package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"

	"dario.cat/mergo"
	"github.com/Kong/ai-deck-converter/convert"
	"github.com/kong/go-database-reconciler/pkg/cprint"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	aiSyncSourceFiles   []string
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

	targetContent, err := buildAiSyncTargetContent(aiSyncSourceFiles)
	if err != nil {
		return err
	}

	injectmanagedByAIDeckTag(targetContent)

	return syncContent(ctx, targetContent, false, aiSyncParallelism, 0,
		aiSyncWorkspace, aiSyncJSONOutput, ApplyTypeFull)
}

// buildAiSyncTargetContent reads every AI Gateway source referenced by
// filenames (files, directories, or "-" for stdin), converts each to a decK
// configuration, and merges the results into a single target content. This
// mirrors how `gateway sync` reads and merges multiple state files.
func buildAiSyncTargetContent(filenames []string) (*file.Content, error) {
	sources, err := readAiSyncSources(filenames)
	if err != nil {
		return nil, err
	}

	var merged file.Content
	for _, src := range sources {
		convertedYAML, warnings, err := convert.Convert(src.content, convert.Options{})
		if err != nil {
			return nil, fmt.Errorf("converting %s: %w", src.name, err)
		}
		reportAiConversionWarnings(warnings, aiSyncJSONOutput)

		content, err := file.GetContentFromReader(bytes.NewReader(convertedYAML), file.EnvVarsSkip)
		if err != nil {
			return nil, fmt.Errorf("parsing converted configuration from %s: %w", src.name, err)
		}
		if err := mergo.Merge(&merged, content, mergo.WithAppendSlice); err != nil {
			return nil, fmt.Errorf("merging converted configuration from %s: %w", src.name, err)
		}
	}

	return &merged, nil
}

// aiSyncSource pairs a source's raw bytes with a human-readable name used in
// error messages ("stdin" or the file path).
type aiSyncSource struct {
	name    string
	content []byte
}

// readAiSyncSources expands each entry in filenames into the AI Gateway source
// documents it refers to. An entry may be "-" (stdin), a single file, or a
// directory (in which case every YAML/JSON config file within it is read),
// matching the file resolution done by `gateway sync`.
func readAiSyncSources(filenames []string) ([]aiSyncSource, error) {
	var sources []aiSyncSource
	for _, fileOrDir := range filenames {
		if fileOrDir == "-" {
			content, err := io.ReadAll(os.Stdin)
			if err != nil {
				return nil, fmt.Errorf("failed to read from stdin: %w", err)
			}
			sources = append(sources, aiSyncSource{name: "stdin", content: content})
			continue
		}

		finfo, err := os.Stat(fileOrDir)
		if err != nil {
			return nil, fmt.Errorf("reading source file: %w", err)
		}

		var files []string
		if finfo.IsDir() {
			files, err = utils.ConfigFilesInDir(fileOrDir)
			if err != nil {
				return nil, fmt.Errorf("getting files from directory: %w", err)
			}
		} else {
			files = []string{fileOrDir}
		}

		for _, f := range files {
			content, err := os.ReadFile(f)
			if err != nil {
				return nil, fmt.Errorf("failed to read source file %s: %w", f, err)
			}
			sources = append(sources, aiSyncSource{name: f, content: content})
		}
	}
	return sources, nil
}

// injectmanagedByAIDeckTag ensures the AI-managed select tag is present in
// _info.select_tags, adding it without discarding any tags already declared.
// This scopes the sync (including pruning) to AI Gateway entities only, the same
// way `ai dump` scopes its read.
func injectmanagedByAIDeckTag(content *file.Content) {
	if content.Info == nil {
		content.Info = &file.Info{}
	}
	if slices.Contains(content.Info.SelectorTags, managedByAIDeckTag) {
		return
	}
	content.Info.SelectorTags = append(content.Info.SelectorTags, managedByAIDeckTag)
}

func newAiSyncCmd() *cobra.Command {
	aiSyncCmd := &cobra.Command{
		Use:   "sync [flags] [ai-gateway-state-files...]",
		Short: "Sync AI Gateway configuration to Kong",
		Long: `The ai sync command reads AI Gateway configuration files and syncs them to Kong AI Gateway,
tagging every managed entity with 'managed_by:deck-ai'.

The AI Gateway state files are provided as positional arguments. Multiple files
and/or directories may be given; directories are searched for YAML and JSON
config files, and the contents of all sources are merged. Use '-' to read from
stdin (the default when no argument is given).

This is the direct equivalent of running 'deck file ai2kong' followed by
'deck gateway sync' on the result.`,
		Args: cobra.MinimumNArgs(0),
		PreRunE: func(_ *cobra.Command, args []string) error {
			aiSyncSourceFiles = args
			if len(aiSyncSourceFiles) == 0 {
				aiSyncSourceFiles = []string{"-"}
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
