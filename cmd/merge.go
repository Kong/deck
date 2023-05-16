/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/kong/go-apiops/deckformat"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/merge"
	"github.com/spf13/cobra"
)

// Executes the CLI command "merge"
func executeMerge(cmd *cobra.Command, args []string) {
	outputFilename, err := cmd.Flags().GetString("output-file")
	if err != nil {
		log.Fatalf("failed getting cli argument 'output-file'; %s", err)
	}

	var asYaml bool
	{
		outputFormat, err := cmd.Flags().GetString("format")
		if err != nil {
			log.Fatalf("failed getting cli argument 'format'; %s", err)
		}
		if outputFormat == "yaml" {
			asYaml = true
		} else if outputFormat == "json" {
			asYaml = false
		} else {
			log.Fatalf("expected '--format' to be either 'yaml'"+
				" or 'json', got: '%s'", outputFormat)
		}
	}

	// do the work: read/merge
	merged, info := merge.MustFiles(args)

	historyEntry := deckformat.HistoryNewEntry("merge")
	historyEntry["output"] = outputFilename
	historyEntry["files"] = info
	deckformat.HistoryClear(merged)
	deckformat.HistoryAppend(merged, historyEntry)

	filebasics.MustWriteSerializedFile(outputFilename, merged, asYaml)
}

//
//
// Define the CLI data for the merge command
//
//

func newMergeCmd() *cobra.Command {
	mergeCmd := &cobra.Command{
		Use:   "merge [flags] filename [...filename]",
		Short: "Merges multiple decK files into one",
		Long: `Merges multiple decK files into one.

The files can be either json or yaml format. Will merge all top-level arrays by simply
concatenating them. Any other keys will be copied. The files will be processed in the order
provided. No checks on content will be done, eg. duplicates, nor any validations.

If the input files are not compatible an error will be returned. Compatibility is
determined by the '_transform' and '_format_version' fields.`,
		Run:  executeMerge,
		Args: cobra.MinimumNArgs(1),
	}

	mergeCmd.Flags().StringP("output-file", "o", "-", "output file to write. Use - to write to stdout")
	mergeCmd.Flags().StringP("format", "", "yaml", "output format: yaml or json")

	return mergeCmd
}
