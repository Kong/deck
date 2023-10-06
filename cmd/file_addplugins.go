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
	"github.com/kong/go-apiops/jsonbasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/kong/go-apiops/plugins"
	"github.com/spf13/cobra"
)

var (
	cmdAddPluginsOverwrite     bool
	cmdAddPluginsInputFilename string
	cmdAddPluginOutputFilename string
	cmdAddPluginOutputFormat   string
	cmdAddPluginsSelectors     []string
	cmdAddPluginsStrConfigs    []string
)

// Executes the CLI command "add-plugins"
func executeAddPlugins(cmd *cobra.Command, cfgFiles []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)

	cmdAddPluginOutputFormat = strings.ToUpper(cmdAddPluginOutputFormat)

	var pluginConfigs []map[string]interface{}
	{
		for _, strConfig := range cmdAddPluginsStrConfigs {
			pluginConfig, err := filebasics.Deserialize([]byte(strConfig))
			if err != nil {
				return fmt.Errorf("failed to deserialize plugin config '%s'; %w", strConfig, err)
			}
			pluginConfigs = append(pluginConfigs, pluginConfig)
		}
	}

	var pluginFiles []plugins.DeckPluginFile
	{
		for _, filename := range cfgFiles {
			var file plugins.DeckPluginFile
			if err := file.ParseFile(filename); err != nil {
				return fmt.Errorf("failed to parse plugin file '%s'; %w", filename, err)
			}
			pluginFiles = append(pluginFiles, file)
		}
	}

	// do the work: read/add-plugins/write
	jsondata, err := filebasics.DeserializeFile(cmdAddPluginsInputFilename)
	if err != nil {
		return fmt.Errorf("failed to read input file '%s'; %w", cmdAddPluginsInputFilename, err)
	}
	yamlNode := jsonbasics.ConvertToYamlNode(jsondata)

	// apply CLI flags
	plugger := plugins.Plugger{}
	plugger.SetYamlData(yamlNode)
	err = plugger.SetSelectors(cmdAddPluginsSelectors)
	if err != nil {
		return fmt.Errorf("failed to set selectors; %w", err)
	}
	err = plugger.AddPlugins(pluginConfigs, cmdAddPluginsOverwrite)
	if err != nil {
		return fmt.Errorf("failed to add plugins; %w", err)
	}
	yamlNode = plugger.GetYamlData()

	// apply plugin-files
	for i, pluginFile := range pluginFiles {
		err = pluginFile.Apply(yamlNode)
		if err != nil {
			return fmt.Errorf("failed to apply plugin file '%s'; %w", cfgFiles[i], err)
		}
	}
	jsondata = plugger.GetData()

	trackInfo := deckformat.HistoryNewEntry("add-plugins")
	trackInfo["input"] = cmdAddPluginsInputFilename
	trackInfo["output"] = cmdAddPluginOutputFilename
	trackInfo["overwrite"] = cmdAddPluginsOverwrite
	if len(pluginConfigs) > 0 {
		trackInfo["configs"] = pluginConfigs
	}
	if len(cfgFiles) > 0 {
		trackInfo["pluginfiles"] = cfgFiles
	}
	trackInfo["selectors"] = cmdAddPluginsSelectors
	deckformat.HistoryAppend(jsondata, trackInfo)

	return filebasics.WriteSerializedFile(
		cmdAddPluginOutputFilename,
		jsondata,
		filebasics.OutputFormat(cmdAddPluginOutputFormat))
}

//
//
// Define the CLI data for the add-plugins command
//
//

func newAddPluginsCmd() *cobra.Command {
	addPluginsCmd := &cobra.Command{
		Use:   "add-plugins [flags] [...plugin-files]",
		Short: "Add plugins to objects in a decK file",
		Long: `Add plugins to objects in a decK file.

The plugins are added to all objects that match the selector expressions. If no
selectors are given, the plugins are added to the top-level 'plugins' array.

The plugin files have the following format (JSON or YAML) and are applied in the
order they are given:

	{ "_format_version": "1.0",
	  "add-plugins": [
	    { "selectors": [
	        "$..services[*]"
	      ],
	      "overwrite": false,
	      "plugins": [
	        { "name": "my-plugin",
	          "config": {
	            "my-property": "value"
	          }
	        }
	      ]
	    }
	  ]
	}
`,
		RunE: executeAddPlugins,
		Args: cobra.MinimumNArgs(0),
	}

	addPluginsCmd.Flags().StringVarP(&cmdAddPluginsInputFilename, "state", "s", "-",
		"decK file to process. Use - to read from stdin.")
	addPluginsCmd.Flags().StringArrayVar(&cmdAddPluginsSelectors, "selector", []string{},
		"JSON path expression to select plugin-owning objects to add plugins to.\n"+
			"Defaults to the top-level (selector '$'). Repeat for multiple selectors.")
	addPluginsCmd.Flags().StringArrayVar(&cmdAddPluginsStrConfigs, "config", []string{},
		"JSON snippet containing the plugin configuration to add. Repeat to add\n"+
			"multiple plugins.")
	addPluginsCmd.Flags().BoolVar(&cmdAddPluginsOverwrite, "overwrite", false,
		"Specify this flag to overwrite plugins by the same name if they already\n"+
			"exist in an array. The default behavior is to skip existing plugins.")
	addPluginsCmd.Flags().StringVarP(&cmdAddPluginOutputFilename, "output-file", "o", "-",
		"Output file to write to. Use - to write to stdout.")
	addPluginsCmd.Flags().StringVarP(&cmdAddPluginOutputFormat, "format", "", string(filebasics.OutputFormatYaml),
		"Output format: "+string(filebasics.OutputFormatJSON)+" or "+string(filebasics.OutputFormatYaml))

	return addPluginsCmd
}
