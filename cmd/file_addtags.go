/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/kong/go-apiops/deckformat"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/kong/go-apiops/tags"
	"github.com/spf13/cobra"
)

var (
	cmdAddTagsInputFilename  string
	cmdAddTagsOutputFilename string
	cmdAddTagsOutputFormat   string
	cmdAddTagsSelectors      []string
)

// Executes the CLI command "add-tags"
func executeAddTags(cmd *cobra.Command, tagsToAdd []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)

	cmdAddTagsOutputFormat = strings.ToUpper(cmdAddTagsOutputFormat)

	// do the work: read/add-tags/write
	data, err := filebasics.DeserializeFile(cmdAddTagsInputFilename)
	if err != nil {
		return fmt.Errorf("failed to read input file '%s'; %w", cmdAddTagsInputFilename, err)
	}

	tagger := tags.Tagger{}
	tagger.SetData(data)
	err = tagger.SetSelectors(cmdAddTagsSelectors)
	if err != nil {
		return fmt.Errorf("failed to set selectors; %w", err)
	}
	err = tagger.AddTags(tagsToAdd)
	if err != nil {
		return fmt.Errorf("failed to add tags; %w", err)
	}
	data = tagger.GetData()

	trackInfo := deckformat.HistoryNewEntry("add-tags")
	trackInfo["input"] = cmdAddTagsInputFilename
	trackInfo["output"] = cmdAddTagsOutputFilename
	trackInfo["tags"] = tagsToAdd
	trackInfo["selectors"] = cmdAddTagsSelectors
	deckformat.HistoryAppend(data, trackInfo)

	return filebasics.WriteSerializedFile(cmdAddTagsOutputFilename, data, filebasics.OutputFormat(cmdAddTagsOutputFormat))
}

//
//
// Define the CLI data for the add-tags command
//
//

func newAddTagsCmd() *cobra.Command {
	addTagsCmd := &cobra.Command{
		Use:   "add-tags [flags] tag [...tag]",
		Short: "Add tags to objects in a decK file",
		Long: `Add tags to objects in a decK file.

The tags are added to all objects that match the selector expressions. If no
selectors are given, all Kong entities are tagged.`,
		RunE: executeAddTags,
		Args: cobra.MinimumNArgs(1),
	}

	addTagsCmd.Flags().StringVarP(&cmdAddTagsInputFilename, "state", "s", "-",
		"decK file to process. Use - to read from stdin.")
	addTagsCmd.Flags().StringArrayVar(&cmdAddTagsSelectors, "selector", []string{},
		"JSON path expression to select objects to add tags to.\n"+
			"Defaults to all Kong entities. Repeat for multiple selectors.")
	addTagsCmd.Flags().StringVarP(&cmdAddTagsOutputFilename, "output-file", "o", "-",
		"Output file to write to. Use - to write to stdout.")
	addTagsCmd.Flags().StringVarP(&cmdAddTagsOutputFormat, "format", "", string(filebasics.OutputFormatYaml),
		"Output format: "+string(filebasics.OutputFormatJSON)+" or "+string(filebasics.OutputFormatYaml))

	return addTagsCmd
}
