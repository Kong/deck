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

// Executes the CLI command "list-tags"
func executeListTags(cmd *cobra.Command, _ []string) error {
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

	// do the work: read/list-tags/write
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
	list, err := tagger.ListTags()
	if err != nil {
		return fmt.Errorf("failed to list tags; %w", err)
	}

	if outputFormat == "PLAIN" {
		// return as a plain text format, unix style; line separated
		result := []byte(strings.Join(list, "\n"))
		return filebasics.WriteFile(outputFilename, result)
	}
	// return as yaml/json, create an object containing only a tags-array
	result := make(map[string]interface{})
	result["tags"] = list
	return filebasics.WriteSerializedFile(outputFilename, result, outputFormat)
}

//
//
// Define the CLI data for the list-tags command
//
//

func newListTagsCmd() *cobra.Command {
	ListTagsCmd := &cobra.Command{
		Use:   "list-tags [flags]",
		Short: "Lists current tags to objects in a decK file",
		Long: `Lists current tags to objects in a decK file.

The tags will be collected from all objects that match the selector expressions. If no
selectors are given, all Kong entities will be scanned.`,
		RunE: executeListTags,
		Args: cobra.NoArgs,
	}

	ListTagsCmd.Flags().StringP("state", "s", "-", "decK file to process. Use - to read from stdin")
	ListTagsCmd.Flags().StringArray("selector", []string{}, "JSON path expression to select "+
		"objects to scan for tags,\ndefaults to all Kong entities (repeat for multiple selectors)")
	ListTagsCmd.Flags().StringP("output-file", "o", "-", "output file to write. Use - to write to stdout")
	ListTagsCmd.Flags().StringP("format", "", "PLAIN", "output format: "+
		filebasics.OutputFormatJSON+", "+filebasics.OutputFormatYaml+", or PLAIN")

	return ListTagsCmd
}
