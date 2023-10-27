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
	cmdO2KskipID         bool
	cmdO2KinsoCompat     bool
)

// Executes the CLI command "openapi2kong"
func executeOpenapi2Kong(cmd *cobra.Command, _ []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)
	_ = sendAnalytics("file-openapi2kong", "", modeLocal)

	if len(cmdO2KentityTags) == 0 {
		cmdO2KentityTags = nil
	}

	cmdO2KoutputFormat = strings.ToUpper(cmdO2KoutputFormat)

	if cmdO2KinsoCompat {
		cmdO2KskipID = true // this is implicit in inso compatibility mode
	}
	options := openapi2kong.O2kOptions{
		Tags:       cmdO2KentityTags,
		DocName:    cmdO2KdocName,
		SkipID:     cmdO2KskipID,
		InsoCompat: cmdO2KinsoCompat,
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
custom annotations (x-kong-... directives).

The output will be targeted at Kong version 3.x.
`,
		RunE: executeOpenapi2Kong,
		Example: "# Convert an OAS file, adding 2 tags, and namespacing the UUIDs to a unique name\n" +
			"cat service_oas.yml | deck file openapi2kong --select-tag=serviceA,teamB --uuid-base=unique-service-name",
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
	openapi2kongCmd.Flags().BoolVar(&cmdO2KskipID, "no-id", false,
		"Setting this flag will skip UUID generation for entities (no 'id' fields\n"+
			"will be added, implicit if '--inso-compatible' is set).")
	openapi2kongCmd.Flags().BoolVarP(&cmdO2KinsoCompat, "inso-compatible", "i", false,
		"This flag will enable Inso compatibility. The generated entity names will be\n"+
			"the same, and no 'id' fields will be gnerated.")

	return openapi2kongCmd
}
