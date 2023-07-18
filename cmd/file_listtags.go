/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/kong/go-apiops/tags"
	"github.com/spf13/cobra"
)

const PlainOutputFormat = "PLAIN"

var (
	cmdListTagsInputFilename  string
	cmdListTagsOutputFilename string
	cmdListTagsOutputFormat   string
	cmdListTagsSelectors      []string
)

// Executes the CLI command "list-tags"
func executeListTags(cmd *cobra.Command, _ []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)

	cmdListTagsOutputFormat = strings.ToUpper(cmdListTagsOutputFormat)

	// do the work: read/list-tags/write
	data, err := filebasics.DeserializeFile(cmdListTagsInputFilename)
	if err != nil {
		return fmt.Errorf("failed to read input file '%s'; %w", cmdListTagsInputFilename, err)
	}

	tagger := tags.Tagger{}
	tagger.SetData(data)
	err = tagger.SetSelectors(cmdListTagsSelectors)
	if err != nil {
		return fmt.Errorf("failed to set selectors; %w", err)
	}
	list, err := tagger.ListTags()
	if err != nil {
		return fmt.Errorf("failed to list tags; %w", err)
	}

	if cmdListTagsOutputFormat == PlainOutputFormat {
		// return as a plain text format, unix style; line separated
		result := []byte(strings.Join(list, "\n"))
		return filebasics.WriteFile(cmdListTagsOutputFilename, result)
	}
	// return as yaml/json, create an object containing only a tags-array
	result := make(map[string]interface{})
	result["tags"] = list
	return filebasics.WriteSerializedFile(
		cmdListTagsOutputFilename,
		result,
		filebasics.OutputFormat(cmdListTagsOutputFormat))
}

//
//
// Define the CLI data for the list-tags command
//
//

func newListTagsCmd() *cobra.Command {
	ListTagsCmd := &cobra.Command{
		Use:   "list-tags [flags]",
		Short: "List current tags from objects in a decK file",
		Long: `List current tags from objects in a decK file.

The tags are collected from all objects that match the selector expressions. If no
selectors are given, all Kong entities will be scanned.`,
		RunE: executeListTags,
		Args: cobra.NoArgs,
	}

	ListTagsCmd.Flags().StringVarP(&cmdListTagsInputFilename, "state", "s", "-",
		"decK file to process. Use - to read from stdin.")
	ListTagsCmd.Flags().StringArrayVar(&cmdListTagsSelectors, "selector", []string{},
		"JSON path expression to select objects to scan for tags.\n"+
			"Defaults to all Kong entities. Repeat for multiple selectors.")
	ListTagsCmd.Flags().StringVarP(&cmdListTagsOutputFilename, "output-file", "o", "-",
		"Output file to write to. Use - to write to stdout.")
	ListTagsCmd.Flags().StringVarP(&cmdListTagsOutputFormat, "format", "", PlainOutputFormat,
		"Output format: "+string(filebasics.OutputFormatJSON)+", "+string(filebasics.OutputFormatYaml)+
			", or "+string(PlainOutputFormat))

	return ListTagsCmd
}
