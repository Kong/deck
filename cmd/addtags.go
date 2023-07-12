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

// Executes the CLI command "add-tags"
func executeAddTags(cmd *cobra.Command, tagsToAdd []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)

	inputFilename, err := cmd.Flags().GetString("state")
	if err != nil {
		return fmt.Errorf("failed getting cli argument 'state'; %w", err)
	}

	outputFilename, err := cmd.Flags().GetString("output-file")
	if err != nil {
		return fmt.Errorf("failed getting cli argument 'output-file'; %w", err)
	}

	var outputFormat string
	{
		outputFormat, err = cmd.Flags().GetString("format")
		if err != nil {
			return fmt.Errorf("failed getting cli argument 'format'; %w", err)
		}
		outputFormat = strings.ToUpper(outputFormat)
	}

	var selectors []string
	{
		selectors, err = cmd.Flags().GetStringArray("selector")
		if err != nil {
			return fmt.Errorf("failed getting cli argument 'selector'; %w", err)
		}
	}

	// do the work: read/add-tags/write
	data, err := filebasics.DeserializeFile(inputFilename)
	if err != nil {
		return fmt.Errorf("failed to read input file '%s'; %w", inputFilename, err)
	}

	tagger := tags.Tagger{}
	tagger.SetData(data)
	err = tagger.SetSelectors(selectors)
	if err != nil {
		return fmt.Errorf("failed to set selectors; %w", err)
	}
	err = tagger.AddTags(tagsToAdd)
	if err != nil {
		return fmt.Errorf("failed to add tags; %w", err)
	}
	data = tagger.GetData()

	trackInfo := deckformat.HistoryNewEntry("add-tags")
	trackInfo["input"] = inputFilename
	trackInfo["output"] = outputFilename
	trackInfo["tags"] = tagsToAdd
	trackInfo["selectors"] = selectors
	deckformat.HistoryAppend(data, trackInfo)

	return filebasics.WriteSerializedFile(outputFilename, data, outputFormat)
}

//
//
// Define the CLI data for the add-tags command
//
//

func newAddTagsCmd() *cobra.Command {
	addTagsCmd := &cobra.Command{
		Use:   "add-tags [flags] tag [...tag]",
		Short: "Adds tags to objects in a decK file",
		Long: `Adds tags to objects in a decK file.

The tags are added to all objects that match the selector expressions. If no
selectors are given, all Kong entities are tagged.`,
		RunE: executeAddTags,
		Args: cobra.MinimumNArgs(1),
	}

	addTagsCmd.Flags().StringP("state", "s", "-", "decK file to process. Use - to read from stdin")
	addTagsCmd.Flags().StringArray("selector", []string{}, "JSON path expression to select "+
		"objects to add tags to,\ndefaults to all Kong entities (repeat for multiple selectors)")
	addTagsCmd.Flags().StringP("output-file", "o", "-", "output file to write. Use - to write to stdout")
	addTagsCmd.Flags().StringP("format", "", filebasics.OutputFormatYaml, "output format: "+
		filebasics.OutputFormatJSON+" or "+filebasics.OutputFormatYaml)

	return addTagsCmd
}
