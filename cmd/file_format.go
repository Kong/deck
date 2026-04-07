package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/kong/go-apiops/deckformat"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/jsonbasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/spf13/cobra"
)

var (
	cmdFileFormatOutputFilename string
	cmdFileFormatOutputFormat   string
)

const (
	fileFormatTypeDeck   = "deck"
	fileFormatTypeDBless = "dbless"
)

// convertDeckToDBless converts a decK format file to DBless format.
// It is the inverse of deckformat.ConvertDBless.
//
// The following transformations are applied:
//   - consumer_groups[*].plugins → top-level consumer_group_plugins (with consumer_group field)
//   - consumers[*].groups → top-level consumer_group_consumers (with consumer and consumer_group fields)
func convertDeckToDBless(data map[string]interface{}) (map[string]interface{}, error) {
	// Step 1: Extract consumer_groups[*].plugins into top-level consumer_group_plugins.
	consumerGroups, err := jsonbasics.GetObjectArrayField(data, "consumer_groups")
	if err != nil {
		return nil, fmt.Errorf("failed to read 'consumer_groups'; %w", err)
	}

	var consumerGroupPlugins []map[string]interface{}
	for i, consumerGroup := range consumerGroups {
		groupName, err := jsonbasics.GetStringField(consumerGroup, "name")
		if err != nil {
			return nil, fmt.Errorf("failed to read 'consumer_groups[%d].name'; %w", i, err)
		}

		plugins, err := jsonbasics.GetObjectArrayField(consumerGroup, "plugins")
		if err != nil {
			return nil, fmt.Errorf("failed to read 'consumer_groups[%d].plugins'; %w", i, err)
		}

		for _, plugin := range plugins {
			plugin["consumer_group"] = groupName
			consumerGroupPlugins = append(consumerGroupPlugins, plugin)
		}
		// Remove nested plugins from the consumer_group entry.
		jsonbasics.SetObjectArrayField(consumerGroup, "plugins", nil)
	}

	if len(consumerGroupPlugins) > 0 {
		jsonbasics.SetObjectArrayField(data, "consumer_group_plugins", consumerGroupPlugins)
	}

	// Step 2: Extract consumers[*].groups into top-level consumer_group_consumers.
	consumers, err := jsonbasics.GetObjectArrayField(data, "consumers")
	if err != nil {
		return nil, fmt.Errorf("failed to read 'consumers'; %w", err)
	}

	var consumerGroupConsumers []map[string]interface{}
	for i, consumer := range consumers {
		username, err := jsonbasics.GetStringField(consumer, "username")
		if err != nil {
			return nil, fmt.Errorf("failed to read 'consumers[%d].username'; %w", i, err)
		}

		groups, err := jsonbasics.GetObjectArrayField(consumer, "groups")
		if err != nil {
			return nil, fmt.Errorf("failed to read 'consumers[%d].groups'; %w", i, err)
		}

		for j, group := range groups {
			groupName, err := jsonbasics.GetStringField(group, "name")
			if err != nil {
				return nil, fmt.Errorf("failed to read 'consumers[%d].groups[%d].name'; %w", i, j, err)
			}
			entry := map[string]interface{}{
				"consumer":       username,
				"consumer_group": groupName,
			}
			consumerGroupConsumers = append(consumerGroupConsumers, entry)
		}
		// Remove nested groups from the consumer entry.
		jsonbasics.SetObjectArrayField(consumer, "groups", nil)
	}

	if len(consumerGroupConsumers) > 0 {
		jsonbasics.SetObjectArrayField(data, "consumer_group_consumers", consumerGroupConsumers)
	}

	return data, nil
}

// executeFileFormat is the handler for the "file format" command.
func executeFileFormat(cmd *cobra.Command, args []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)
	_ = sendAnalytics("file-format", "", modeLocal)

	cmdFileFormatOutputFormat = strings.ToUpper(cmdFileFormatOutputFormat)

	formatType := args[0]
	inputFilename := args[1]

	data, err := filebasics.DeserializeFile(inputFilename)
	if err != nil {
		return fmt.Errorf("failed to read input file '%s'; %w", inputFilename, err)
	}

	switch formatType {
	case fileFormatTypeDeck:
		data, err = deckformat.ConvertDBless(data)
		if err != nil {
			return fmt.Errorf("failed to convert DBless to decK format; %w", err)
		}
	case fileFormatTypeDBless:
		data, err = convertDeckToDBless(data)
		if err != nil {
			return fmt.Errorf("failed to convert decK to DBless format; %w", err)
		}
	}

	trackInfo := deckformat.HistoryNewEntry("format")
	trackInfo["input"] = inputFilename
	trackInfo["output"] = cmdFileFormatOutputFilename
	trackInfo["type"] = formatType
	deckformat.HistoryAppend(data, trackInfo)

	return filebasics.WriteSerializedFile(
		cmdFileFormatOutputFilename,
		data,
		filebasics.OutputFormat(cmdFileFormatOutputFormat))
}

// newFileFormatCmd returns the cobra command for "deck file format".
func newFileFormatCmd() *cobra.Command {
	formatCmd := &cobra.Command{
		Use:   fmt.Sprintf("format [flags] %s|%s filename", fileFormatTypeDeck, fileFormatTypeDBless),
		Short: "Convert between decK and DBless file formats",
		Long: `Convert Kong configuration files between decK and Kong DBless formats.

The two formats differ in how consumer-group related entities are represented:
  - decK format:   consumer group plugins are nested under consumer_groups[*].plugins,
                   and consumer group memberships are nested under consumers[*].groups.
  - DBless format: consumer group plugins are stored in a top-level consumer_group_plugins
                   array, and memberships are stored in a top-level consumer_group_consumers
                   array.

Use 'deck' as the type to convert a DBless file into decK format.
Use 'dbless' as the type to convert a decK file into DBless format.`,
		Example: "# Convert a DBless file to decK format\n" +
			"deck file format deck dbless.yaml\n\n" +
			"# Convert a decK file to DBless format\n" +
			"deck file format dbless deck.yaml",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return err
			}
			validTypes := []string{fileFormatTypeDeck, fileFormatTypeDBless}
			return validateInputFlag("type", args[0], validTypes, "")
		},
		RunE: executeFileFormat,
	}

	formatCmd.Flags().StringVarP(&cmdFileFormatOutputFilename, "output-file", "o", "-",
		"Output file to write to. Use - to write to stdout.")
	formatCmd.Flags().StringVar(&cmdFileFormatOutputFormat, "format", "yaml",
		"Output file format: yaml or json.")

	return formatCmd
}
