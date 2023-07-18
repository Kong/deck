package cmd

import (
	"log"

	"github.com/kong/go-apiops/deckformat"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/kong/go-apiops/merge"
	"github.com/spf13/cobra"
)

var (
	cmdMergeOutputFilename string
	cmdMergeOutputFormat   string
)

// Executes the CLI command "merge"
func executeMerge(cmd *cobra.Command, args []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)

	// do the work: read/merge
	merged, info, err := merge.Files(args)
	if err != nil {
		return err
	}

	historyEntry := deckformat.HistoryNewEntry("merge")
	historyEntry["output"] = cmdMergeOutputFilename
	historyEntry["files"] = info
	deckformat.HistoryClear(merged)
	deckformat.HistoryAppend(merged, historyEntry)

	return filebasics.WriteSerializedFile(
		cmdMergeOutputFilename,
		merged,
		filebasics.OutputFormat(cmdMergeOutputFormat))
}

//
//
// Define the CLI data for the merge command
//
//

func newMergeCmd() *cobra.Command {
	mergeCmd := &cobra.Command{
		Use:   "merge [flags] filename [...filename]",
		Short: "Merge multiple decK files into one",
		Long: `Merge multiple decK files into one.

The files can be in either JSON or YAML format. Merges all top-level arrays by
concatenating them. Any other keys are copied. The files are processed in the order
provided. 

Doesn't perform any checks on content, e.g. duplicates, or any validations.

If the input files are not compatible, returns an error. Compatibility is
determined by the '_transform' and '_format_version' fields.`,
		RunE: executeMerge,
		Args: cobra.MinimumNArgs(1),
	}

	mergeCmd.Flags().StringVarP(&cmdMergeOutputFilename, "output-file", "o", "-",
		"Output file to write to. Use - to write to stdout.")
	mergeCmd.Flags().StringVarP(&cmdMergeOutputFormat, "format", "", "yaml", "output format: yaml or json")

	return mergeCmd
}
