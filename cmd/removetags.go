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
	cmdRemoveTagsKeepEmptyArrays bool
	cmdRemoveTagsKeepOnlyTags    bool
	cmdRemoveTagsInputFilename   string
	cmdRemoveTagsOutputFilename  string
	cmdRemoveTagsOutputFormat    string
	cmdRemoveTagsSelectors       []string
)

// Executes the CLI command "remove-tags"
func executeRemoveTags(cmd *cobra.Command, tagsToRemove []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)

	cmdRemoveTagsOutputFormat = strings.ToUpper(cmdRemoveTagsOutputFormat)

	if !cmdRemoveTagsKeepOnlyTags && len(tagsToRemove) == 0 {
		return fmt.Errorf("no tags to remove")
	}

	// do the work: read/remove-tags/write
	data, err := filebasics.DeserializeFile(cmdRemoveTagsInputFilename)
	if err != nil {
		return fmt.Errorf("failed to read input file '%s'; %w", cmdRemoveTagsInputFilename, err)
	}

	tagger := tags.Tagger{}
	tagger.SetData(data)
	err = tagger.SetSelectors(cmdRemoveTagsSelectors)
	if err != nil {
		return fmt.Errorf("failed to set selectors; %w", err)
	}
	if cmdRemoveTagsKeepOnlyTags {
		err = tagger.RemoveUnknownTags(tagsToRemove, !cmdRemoveTagsKeepEmptyArrays)
	} else {
		err = tagger.RemoveTags(tagsToRemove, !cmdRemoveTagsKeepEmptyArrays)
	}
	if err != nil {
		return fmt.Errorf("failed to remove tags; %w", err)
	}
	data = tagger.GetData()

	trackInfo := deckformat.HistoryNewEntry("remove-tags")
	trackInfo["input"] = cmdRemoveTagsInputFilename
	trackInfo["output"] = cmdRemoveTagsOutputFilename
	trackInfo["tags"] = tagsToRemove
	trackInfo["keep-empty-array"] = cmdRemoveTagsKeepEmptyArrays
	trackInfo["selectors"] = cmdRemoveTagsSelectors
	deckformat.HistoryAppend(data, trackInfo)

	return filebasics.WriteSerializedFile(cmdRemoveTagsOutputFilename, data, cmdRemoveTagsOutputFormat)
}

//
//
// Define the CLI data for the remove-tags command
//
//

func newRemoveTagsCmd() *cobra.Command {
	removeTagsCmd := &cobra.Command{
		Use:   "remove-tags [flags] tag [...tag]",
		Short: "Removes tags from objects in a decK file",
		Long: `Removes tags from objects in a decK file.

The listed tags are removed from all objects that match the selector expressions.
If no selectors are given, all Kong entities will be selected.`,
		RunE: executeRemoveTags,
		Example: "  # clear tags 'tag1' and 'tag2' from all services in file 'kong.yml'\n" +
			"  cat kong.yml | go-apiops remove-tags --selector='services[*]' tag1 tag2\n" +
			"\n" +
			"  # clear all tags except 'tag1' and 'tag2' from the file 'kong.yml'\n" +
			"  cat kong.yml | go-apiops remove-tags --keep-only tag1 tag2",
	}

	removeTagsCmd.Flags().BoolVar(&cmdRemoveTagsKeepEmptyArrays, "keep-empty-array", false,
		"keep empty tag-arrays in output")
	removeTagsCmd.Flags().BoolVar(&cmdRemoveTagsKeepOnlyTags, "keep-only", false,
		"setting this flag will remove all tags except the ones listed\n"+
			"(if none are listed, all tags will be removed)")
	removeTagsCmd.Flags().StringVarP(&cmdRemoveTagsInputFilename, "state", "s", "-",
		"decK file to process. Use - to read from stdin")
	removeTagsCmd.Flags().StringArrayVar(&cmdRemoveTagsSelectors, "selector", []string{},
		"JSON path expression to select objects to remove tags from,\n"+
			"defaults to all Kong entities (repeat for multiple selectors)")
	removeTagsCmd.Flags().StringVarP(&cmdRemoveTagsOutputFilename, "output-file", "o", "-",
		"output file to write. Use - to write to stdout")
	removeTagsCmd.Flags().StringVarP(&cmdRemoveTagsOutputFormat, "format", "", filebasics.OutputFormatYaml,
		"output format: "+filebasics.OutputFormatJSON+" or "+filebasics.OutputFormatYaml)

	return removeTagsCmd
}
