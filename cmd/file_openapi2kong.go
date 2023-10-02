package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/kong/go-apiops/deckformat"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/kong/go-apiops/openapi2kong"
	"github.com/spf13/cobra"
)

var (
	cmdO2KinputFilename  string
	cmdO2KoutputFilename string
	cmdO2KdocName        string
	cmdO2KoutputFormat   string
	cmdO2KentityTags     []string
)

// Executes the CLI command "openapi2kong"
func executeOpenapi2Kong(cmd *cobra.Command, _ []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)

	if len(cmdO2KentityTags) == 0 {
		cmdO2KentityTags = nil
	}

	cmdO2KoutputFormat = strings.ToUpper(cmdO2KoutputFormat)

	options := openapi2kong.O2kOptions{
		Tags:    cmdO2KentityTags,
		DocName: cmdO2KdocName,
	}

	trackInfo := deckformat.HistoryNewEntry("openapi2kong")
	trackInfo["input"] = cmdO2KinputFilename
	trackInfo["output"] = cmdO2KoutputFilename
	trackInfo["uuid-base"] = cmdO2KdocName

	// do the work: read/convert/write
	content, err := filebasics.ReadFile(cmdO2KinputFilename)
	if err != nil {
		return err
	}
	result, err := openapi2kong.Convert(content, options)
	if err != nil {
		return fmt.Errorf("failed converting OpenAPI spec '%s'; %w", cmdO2KinputFilename, err)
	}
	deckformat.HistoryAppend(result, trackInfo)
	return filebasics.WriteSerializedFile(cmdO2KoutputFilename, result, filebasics.OutputFormat(cmdO2KoutputFormat))
}

//
//
// Define the CLI data for the openapi2kong command
//
//

func newOpenapi2KongCmd() *cobra.Command {
	openapi2kongCmd := &cobra.Command{
		Use:   "openapi2kong",
		Short: "Convert OpenAPI files to Kong's decK format",
		Long: `Convert OpenAPI files to Kong's decK format.

The example file at https://github.com/Kong/go-apiops/blob/main/docs/learnservice_oas.yaml
has extensive annotations explaining the conversion process, as well as all supported 
custom annotations (x-kong-... directives).`,
		RunE: executeOpenapi2Kong,
		Args: cobra.NoArgs,
	}

	openapi2kongCmd.Flags().StringVarP(&cmdO2KinputFilename, "spec", "s", "-",
		"OpenAPI spec file to process. Use - to read from stdin.")
	openapi2kongCmd.Flags().StringVarP(&cmdO2KoutputFilename, "output-file", "o", "-",
		"Output file to write. Use - to write to stdout.")
	openapi2kongCmd.Flags().StringVarP(&cmdO2KoutputFormat, "format", "", "yaml", "output format: yaml or json")
	openapi2kongCmd.Flags().StringVarP(&cmdO2KdocName, "uuid-base", "", "",
		"The unique base-string for uuid-v5 generation of entity IDs. If omitted,\n"+
			"uses the root-level \"x-kong-name\" directive, or falls back to 'info.title'.)")
	openapi2kongCmd.Flags().StringSliceVar(&cmdO2KentityTags, "select-tag", nil,
		"Select tags to apply to all entities. If omitted, uses the \"x-kong-tags\"\n"+
			"directive from the file.")

	return openapi2kongCmd
}
