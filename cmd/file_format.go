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
//   - plugins[*].partials → top-level plugins_partials (with plugin/partial/path fields)
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

	// Step 3: Extract plugins[*].partials into top-level plugins_partials.
	plugins, err := jsonbasics.GetObjectArrayField(data, "plugins")
	if err != nil {
		return nil, fmt.Errorf("failed to read 'plugins'; %w", err)
	}

	var pluginPartials []map[string]interface{}
	for i, plugin := range plugins {
		pluginRef, err := jsonbasics.GetStringField(plugin, "id")
		if err != nil {
			pluginRef, err = jsonbasics.GetStringField(plugin, "name")
			if err != nil {
				return nil, fmt.Errorf("failed to read 'plugins[%d].id' or 'plugins[%d].name'; %w", i, i, err)
			}
		}

		partials, err := jsonbasics.GetObjectArrayField(plugin, "partials")
		if err != nil {
			return nil, fmt.Errorf("failed to read 'plugins[%d].partials'; %w", i, err)
		}

		for j, partial := range partials {
			partialRef, err := jsonbasics.GetStringField(partial, "id")
			if err != nil {
				partialRef, err = jsonbasics.GetStringField(partial, "name")
				if err != nil {
					return nil, fmt.Errorf(
						"failed to read 'plugins[%d].partials[%d].id' or 'plugins[%d].partials[%d].name'; %w",
						i, j, i, j, err,
					)
				}
			}

			entry := map[string]interface{}{
				"plugin":  pluginRef,
				"partial": partialRef,
			}

			if path, err := jsonbasics.GetStringField(partial, "path"); err == nil {
				entry["path"] = path
			}

			pluginPartials = append(pluginPartials, entry)
		}

		// Remove nested partials from the plugin entry.
		jsonbasics.SetObjectArrayField(plugin, "partials", nil)
	}

	if len(pluginPartials) > 0 {
		jsonbasics.SetObjectArrayField(data, "plugins_partials", pluginPartials)
	}

	return data, nil
}

func convertDBlessToDeck(data map[string]interface{}) (map[string]interface{}, error) {
	converted, err := deckformat.ConvertDBless(data)
	if err != nil {
		return nil, err
	}

	pluginsPartials, err := jsonbasics.GetObjectArrayField(converted, "plugins_partials")
	if err != nil {
		return nil, fmt.Errorf("failed to read 'plugins_partials'; %w", err)
	}
	if len(pluginsPartials) == 0 {
		return converted, nil
	}

	plugins, err := jsonbasics.GetObjectArrayField(converted, "plugins")
	if err != nil {
		return nil, fmt.Errorf("failed to read 'plugins'; %w", err)
	}

	partials, err := jsonbasics.GetObjectArrayField(converted, "partials")
	if err != nil {
		return nil, fmt.Errorf("failed to read 'partials'; %w", err)
	}

	partialsByID := make(map[string]map[string]interface{})
	partialsByName := make(map[string]map[string]interface{})
	for _, partial := range partials {
		if id, err := jsonbasics.GetStringField(partial, "id"); err == nil {
			partialsByID[id] = partial
		}
		if name, err := jsonbasics.GetStringField(partial, "name"); err == nil {
			partialsByName[name] = partial
		}
	}

	findPlugin := func(ref string) map[string]interface{} {
		for _, plugin := range plugins {
			if id, err := jsonbasics.GetStringField(plugin, "id"); err == nil && id == ref {
				return plugin
			}
			if name, err := jsonbasics.GetStringField(plugin, "name"); err == nil && name == ref {
				return plugin
			}
		}
		return nil
	}

	for i, pluginPartial := range pluginsPartials {
		pluginRef, err := jsonbasics.GetStringField(pluginPartial, "plugin")
		if err != nil {
			return nil, fmt.Errorf("failed to read 'plugins_partials[%d].plugin'; %w", i, err)
		}
		plugin := findPlugin(pluginRef)
		if plugin == nil {
			return nil, fmt.Errorf("failed to resolve 'plugins_partials[%d].plugin'='%s' to a plugin in the file", i, pluginRef)
		}

		partialRef, err := jsonbasics.GetStringField(pluginPartial, "partial")
		if err != nil {
			return nil, fmt.Errorf("failed to read 'plugins_partials[%d].partial'; %w", i, err)
		}

		partialEntry := map[string]interface{}{}
		if partial, ok := partialsByID[partialRef]; ok {
			if id, err := jsonbasics.GetStringField(partial, "id"); err == nil {
				partialEntry["id"] = id
			}
			if name, err := jsonbasics.GetStringField(partial, "name"); err == nil {
				partialEntry["name"] = name
			}
		} else if partial, ok := partialsByName[partialRef]; ok {
			if id, err := jsonbasics.GetStringField(partial, "id"); err == nil {
				partialEntry["id"] = id
			}
			if name, err := jsonbasics.GetStringField(partial, "name"); err == nil {
				partialEntry["name"] = name
			}
		} else {
			partialEntry["id"] = partialRef
		}

		if path, err := jsonbasics.GetStringField(pluginPartial, "path"); err == nil {
			partialEntry["path"] = path
		}

		pluginPartialsEntry, err := jsonbasics.GetObjectArrayField(plugin, "partials")
		if err != nil {
			return nil, fmt.Errorf("failed to read 'plugins[%d].partials'; %w", i, err)
		}
		pluginPartialsEntry = append(pluginPartialsEntry, partialEntry)
		jsonbasics.SetObjectArrayField(plugin, "partials", pluginPartialsEntry)
	}

	jsonbasics.SetObjectArrayField(converted, "plugins_partials", nil)
	return converted, nil
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
		data, err = convertDBlessToDeck(data)
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
                   consumer group memberships are nested under consumers[*].groups,
                   and plugin partial links are nested under plugins[*].partials.
  - DBless format: consumer group plugins are stored in a top-level consumer_group_plugins
                   array, memberships are stored in a top-level consumer_group_consumers
                   array, and plugin partial links are stored in top-level plugins_partials.

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
