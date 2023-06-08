package cmd

import (
	"fmt"
	"log"

	"github.com/kong/go-apiops/deckformat"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/jsonbasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/kong/go-apiops/patch"
	"github.com/spf13/cobra"
)

// Executes the CLI command "patch"
func executePatch(cmd *cobra.Command, args []string) error {
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

	var asYaml bool
	{
		outputFormat, err := cmd.Flags().GetString("format")
		if err != nil {
			return fmt.Errorf("failed getting cli argument 'format'; %w", err)
		}
		if outputFormat == "yaml" {
			asYaml = true
		} else if outputFormat == "json" {
			asYaml = false
		} else {
			return fmt.Errorf("expected '--format' to be either 'yaml' or 'json', got: '%s'",
				outputFormat)
		}
	}

	var valuesPatch patch.DeckPatch
	{
		values, err := cmd.Flags().GetStringArray("value")
		if err != nil {
			return fmt.Errorf("failed to retrieve '--value' entries; %w", err)
		}
		valuesPatch.Values, valuesPatch.Remove, err = patch.ValidateValuesFlags(values)
		if err != nil {
			return fmt.Errorf("failed parsing '--value' entry; %w", err)
		}
	}

	{
		s, err := cmd.Flags().GetString("selector")
		if err != nil {
			return fmt.Errorf("failed to retrieve '--selector' entry; %w", err)
		}
		valuesPatch.SelectorSource = s
	}

	patchFiles := make([]patch.DeckPatchFile, 0)
	{
		for _, filename := range args {
			var patchfile patch.DeckPatchFile
			err := patchfile.ParseFile(filename)
			if err != nil {
				return fmt.Errorf("failed to parse '%s': %w", filename, err)
			}
			patchFiles = append(patchFiles, patchfile)
		}
	}

	trackInfo := deckformat.HistoryNewEntry("patch")
	trackInfo["input"] = inputFilename
	trackInfo["output"] = outputFilename
	if len(valuesPatch.Values) != 0 || len(valuesPatch.Remove) != 0 {
		trackInfo["selector"] = valuesPatch.SelectorSource
	}
	if len(valuesPatch.Values) != 0 {
		trackInfo["values"] = valuesPatch.Values
	}
	if len(valuesPatch.Remove) != 0 {
		trackInfo["remove"] = valuesPatch.Remove
	}
	if len(args) != 0 {
		trackInfo["patchfiles"] = args
	}

	// do the work; read/patch/write
	data, err := filebasics.DeserializeFile(inputFilename)
	if err != nil {
		return fmt.Errorf("failed to read input file '%s'; %w", inputFilename, err)
	}
	deckformat.HistoryAppend(data, trackInfo) // add before patching, so patch can operate on it

	yamlNode := jsonbasics.ConvertToYamlNode(data)

	if (len(valuesPatch.Values) + len(valuesPatch.Remove)) > 0 {
		// apply selector + value flags
		logbasics.Debug("applying value-flags")
		err = valuesPatch.ApplyToNodes(yamlNode)
		if err != nil {
			return fmt.Errorf("failed to apply command-line values; %w", err)
		}
	}

	if len(args) > 0 {
		// apply patch files
		for i, patchFile := range patchFiles {
			logbasics.Debug("applying patch-file", "file", i)
			err := patchFile.Apply(yamlNode)
			if err != nil {
				return fmt.Errorf("failed to apply patch-file '%s'; %w", args[i], err)
			}
		}
	}

	data = jsonbasics.ConvertToJSONobject(yamlNode)

	return filebasics.WriteSerializedFile(outputFilename, data, asYaml)
}

//
//
// Define the CLI data for the patch command
//
//

func newPatchCmd() *cobra.Command {
	patchCmd := &cobra.Command{
		Use:   "patch [flags] [...patch-files]",
		Short: "Applies patches on top of a decK file",
		Long: `Applies patches on top of a decK file.

The input file will be read, the patches will be applied, and if successful, written
to the output file. The patches can be specified by a '--selector' and one or more
'--value' tags, or via patch-files.

When using '--selector' and '--values', the items will be selected by the 'selector' which is
a JSONpath query. From the array of nodes found, only the objects will be updated.
The 'values' will be applied on each of the JSONobjects returned by the 'selector'.

The value part must be a valid JSON snippet, so make sure to use single/double quotes
appropriately. If the value is empty, the field will be removed from the object.
Examples:
  --selector="$..services[*]" --value="read_timeout:10000"
  --selector="$..services[*]" --value='_comment:"comment injected by patching"'
  --selector="$..services[*]" --value='_ignore:["ignore1","ignore2"]'
  --selector="$..services[*]" --value='_ignore:' --value='_comment:'

The patchfiles have the following format (JSON or Yaml) and can contain multiple
patches that will be applied in order;

  { "_format_version": "1.0",
    "patches": [
      { "selector": "$..services[*]",
        "values": {
          "read_timeout": 10000,
          "_comment": "comment injected by patching"
        },
        "remove": [ "_ignore" ]
      }
    ]
  }
`,
		RunE: executePatch,
	}

	patchCmd.Flags().StringP("state", "s", "-", "decK file to process. Use - to read from stdin")
	patchCmd.Flags().StringP("output-file", "o", "-", "output file to write. Use - to write to stdout")
	patchCmd.Flags().StringP("format", "", "yaml", "output format: yaml or json")
	patchCmd.Flags().StringP("selector", "", "", "json-pointer identifying element to patch")
	patchCmd.Flags().StringArrayP("value", "", []string{}, "a value to set in the selected entry in "+
		"format <key:value> (can be specified more than once)")
	patchCmd.MarkFlagsRequiredTogether("selector", "value")

	return patchCmd
}
