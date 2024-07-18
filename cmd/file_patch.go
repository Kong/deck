package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/kong/go-apiops/deckformat"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/jsonbasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/kong/go-apiops/patch"
	"github.com/spf13/cobra"
)

var (
	cmdPatchInputFilename  string
	cmdPatchOutputFilename string
	cmdPatchOutputFormat   string
	cmdPatchValues         []string
	cmdPatchSelectors      []string
)

// Executes the CLI command "patch"
func executePatch(cmd *cobra.Command, args []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)
	_ = sendAnalytics("file-patch", "", modeLocal)

	cmdPatchOutputFormat = strings.ToUpper(cmdPatchOutputFormat)

	var valuesPatch patch.DeckPatch
	{
		var err error
		valuesPatch.SelectorSources = cmdPatchSelectors
		valuesPatch.ObjValues, valuesPatch.Remove, valuesPatch.ArrValues, err = patch.ValidateValuesFlags(cmdPatchValues)
		if err != nil {
			return fmt.Errorf("failed parsing '--value' entry; %w", err)
		}
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
	trackInfo["input"] = cmdPatchInputFilename
	trackInfo["output"] = cmdPatchOutputFilename
	if (len(valuesPatch.ObjValues) + len(valuesPatch.Remove) + len(valuesPatch.ArrValues)) > 0 {
		trackInfo["selector"] = valuesPatch.SelectorSources
	}
	if len(valuesPatch.ObjValues) != 0 {
		trackInfo["values"] = valuesPatch.ObjValues
	}
	if len(valuesPatch.ArrValues) != 0 {
		trackInfo["values"] = valuesPatch.ArrValues
	}
	if len(valuesPatch.Remove) != 0 {
		trackInfo["remove"] = valuesPatch.Remove
	}
	if len(args) != 0 {
		trackInfo["patchfiles"] = args
	}

	// do the work; read/patch/write
	data, err := filebasics.DeserializeFile(cmdPatchInputFilename)
	if err != nil {
		return fmt.Errorf("failed to read input file '%s'; %w", cmdPatchInputFilename, err)
	}
	deckformat.HistoryAppend(data, trackInfo) // add before patching, so patch can operate on it

	yamlNode := jsonbasics.ConvertToYamlNode(data)

	if (len(valuesPatch.ObjValues) + len(valuesPatch.Remove) + len(valuesPatch.ArrValues)) > 0 {
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

	return filebasics.WriteSerializedFile(cmdPatchOutputFilename, data, filebasics.OutputFormat(cmdPatchOutputFormat))
}

//
//
// Define the CLI data for the patch command
//
//

func newPatchCmd() *cobra.Command {
	patchCmd := &cobra.Command{
		Use:   "patch [flags] [...patch-files]",
		Short: "Apply patches on top of a decK file",
		Long: `Apply patches on top of a decK file.

The input file is read, the patches are applied, and if successful, written
to the output file. The patches can be specified by a '--selector' and one or more
'--value' tags, or via patch files.

When using '--selector' and '--values', the items are selected by the 'selector', 
which is a JSONpath query. The 'field values' (in '<key:value>' format) are applied on 
each of the JSONObjects returned by the 'selector'. The 'array values' (in 
'[val1, val2]' format) are appended to each of the JSONArrays returned by the 'selector'.

The field values must be a valid JSON snippet, so use single/double quotes
appropriately. If the value is empty, the field is removed from the object.

Examples of valid values:

  # set field "read_timeout" to a numeric value of 10000
	--selector="$..services[*]" --value="read_timeout:10000"

	# set field "_comment" to a string value
	--selector="$..services[*]" --value='_comment:"comment injected by patching"'

	# set field "_ignore" to an array of strings
	--selector="$..services[*]" --value='_ignore:["ignore1","ignore2"]'

	# remove fields "_ignore" and "_comment" from the object
	--selector="$..services[*]" --value='_ignore:' --value='_comment:'

	# append entries to the methods array of all route objects
	--selector="$..routes[*].methods" --value='["OPTIONS"]'


Patch files have the following format (JSON or YAML) and can contain multiple
patches that are applied in order:

	{ "_format_version": "1.0",
	  "patches": [
	    { "selectors": [
	        "$..services[*]"
	      ],
	      "values": {
	        "read_timeout": 10000,
	        "_comment": "comment injected by patching"
	      },
	      "remove": [ "_ignore" ]
	    }
	  ]
	}

If the 'values' object instead is an array, then any arrays returned by the selectors
will get the 'values' appended to them.
`,
		RunE: executePatch,
		PersistentPreRunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 && (len(cmdPatchSelectors) > 0 || len(cmdPatchValues) > 0) {
				return fmt.Errorf("cannot use patch file argument along with '--selector' and '--value'")
			}

			if len(args) == 0 && len(cmdPatchSelectors) == 0 && len(cmdPatchValues) == 0 {
				return fmt.Errorf("must specify at least one of these: " +
					"a patch file argument or a '--selector' and '--value' combination")
			}

			return nil
		},
		Example: "# update the read-timeout on all services\n" +
			"cat kong.yml | deck file patch --selector=\"$..services[*]\" --value=\"read_timeout:10000\"",
	}

	patchCmd.Flags().StringVarP(&cmdPatchInputFilename, "state", "s", "-",
		"decK file to process. Use - to read from stdin.")
	patchCmd.Flags().StringVarP(&cmdPatchOutputFilename, "output-file", "o", "-",
		"Output file to write. Use - to write to stdout.")
	patchCmd.Flags().StringVarP(&cmdPatchOutputFormat, "format", "", "yaml",
		"Output format: yaml or json.")
	patchCmd.Flags().StringArrayVarP(&cmdPatchSelectors, "selector", "", []string{},
		"json-pointer identifying element to patch. Repeat for multiple selectors.)")
	patchCmd.Flags().StringArrayVarP(&cmdPatchValues, "value", "", []string{},
		"A value to set in the selected entry in <key:value> format. Can be specified multiple times.")
	patchCmd.MarkFlagsRequiredTogether("selector", "value")

	return patchCmd
}
