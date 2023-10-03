package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/kong/deck/file"
	"github.com/kong/go-apiops/logbasics"
	"github.com/spf13/cobra"
)

var (
	cmdKong2KicInputFilename  string
	cmdKong2KicOutputFilename string
	//cmdKong2KicApi            string
	cmdKong2KicOutputFormat   string
	cmdKong2KicManifestStyle  string
)

// Executes the CLI command "kong2kic"
func executeKong2Kic(cmd *cobra.Command, _ []string) error {

	var (
		outputContent    *file.Content
		err              error
		outputFileFormat file.Format
	)

	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)

	inputContent, err := file.GetContentFromFiles([]string{cmdKong2KicInputFilename}, false)
	if err != nil {
		return fmt.Errorf("failed reding input file '%s'; %w", cmdKong2KicInputFilename, err)
	}

	outputContent = inputContent.DeepCopy()
	if strings.ToUpper(cmdKong2KicOutputFormat) == "JSON" && strings.ToUpper(cmdKong2KicManifestStyle) == "CRD" {
		outputFileFormat = file.KIC_JSON_CRD
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == "JSON" && strings.ToUpper(cmdKong2KicManifestStyle) == "ANNOTATION" {
		outputFileFormat = file.KIC_JSON_ANNOTATION
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == "YAML" && strings.ToUpper(cmdKong2KicManifestStyle) == "CRD" {
		outputFileFormat = file.KIC_YAML_CRD
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == "YAML" && strings.ToUpper(cmdKong2KicManifestStyle) == "ANNOTATION" {
		outputFileFormat = file.KIC_YAML_ANNOTATION
	} else {
		return fmt.Errorf("invalid combination of output format and manifest style")
	}

	err = file.WriteContentToFile(outputContent, cmdKong2KicOutputFilename, outputFileFormat)

	if err != nil {
		return fmt.Errorf("failed converting Kong to Ingress '%s'; %w", cmdKong2KicInputFilename, err)
	}

	return nil
}

//
//
// Define the CLI data for the openapi2kong command
//
//

func newKong2KicCmd() *cobra.Command {
	kong2KicCmd := &cobra.Command{
		Use:   "kong2kic",
		Short: "Convert Kong configuration files to Kong Ingress Controller (KIC) manifests",
		Long: `Convert Kong configuration files to Kong Ingress Controller (KIC) manifests.
		
Manifests can be generated using annotations in Ingress and Service objects (recommended) or
using the KongIngress CRD. Output in YAML or JSON format.`,
		RunE: executeKong2Kic, 
		Args: cobra.NoArgs,
	}

	kong2KicCmd.Flags().StringVarP(&cmdKong2KicInputFilename, "input-file", "i", "-",
		"Kong spec file to process. Use - to read from stdin.")
	kong2KicCmd.Flags().StringVarP(&cmdKong2KicOutputFilename, "output-file", "o", "-",
		"Output file to write. Use - to write to stdout.")
	kong2KicCmd.Flags().StringVar(&cmdKong2KicManifestStyle, "style", "annotation",
		"Generate manifests with annotations in Service and Ingress, or using the KongIngress CRD: annotation or crd.")
	kong2KicCmd.Flags().StringVarP(&cmdKong2KicOutputFormat, "format", "f", "yaml",
		"output file format: json or yaml.")
	//kong2KicCmd.Flags().StringVarP(&cmdKong2KicApi, "api", "a", "ingress", "[ingress|gateway]")

	return kong2KicCmd
}
