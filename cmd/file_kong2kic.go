package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/kong/deck/kong2kic"
	"github.com/kong/go-apiops/logbasics"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/spf13/cobra"
)

var (
	cmdKong2KicInputFilename  string
	cmdKong2KicOutputFilename string
	cmdKong2KicOutputFormat   string
	cmdKong2KicClassName      string
	cmdKong2KicIngress        bool
	cmdKong2KicVersion        string
)

// Executes the CLI command "kong2kic"
func executeKong2Kic(cmd *cobra.Command, _ []string) error {
	_ = sendAnalytics("file-kong2kic", "", modeLocal)
	var (
		outputContent    *file.Content
		err              error
		outputFileFormat file.Format
		yamlOrJSON       string
	)

	kong2kic.ClassName = cmdKong2KicClassName
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)

	inputContent, err := file.GetContentFromFiles([]string{cmdKong2KicInputFilename}, false)
	if err != nil {
		return fmt.Errorf("failed reading input file '%s'; %w", cmdKong2KicInputFilename, err)
	}

	outputFileFormat, yamlOrJSON, err = validateInput(cmdKong2KicOutputFormat)
	if err != nil {
		return err
	}

	outputContent = inputContent.DeepCopy()
	err = kong2kic.WriteContentToFile(outputContent, cmdKong2KicOutputFilename, outputFileFormat, yamlOrJSON)
	if err != nil {
		return fmt.Errorf("failed converting Kong to Ingress '%s'; %w", cmdKong2KicInputFilename, err)
	}

	return nil
}

func validateInput(cmdKong2KicOutputFormat string) (
	outputFileFormat file.Format,
	yamlOrJSON string,
	err error,
) {
	outputFormat := strings.ToUpper(cmdKong2KicOutputFormat)
	version := cmdKong2KicVersion
	ingress := cmdKong2KicIngress

	// if cmdKong2KicVersion is not 2 or 3 set an error
	if version != "2" && version != "3" {
		err = fmt.Errorf("invalid KIC version '%s'. Please use --help for more information", version)
	} else {
		switch {
		case version == "3" && !ingress:
			outputFileFormat = kong2kic.KICV3GATEWAY
		case version == "3" && ingress:
			outputFileFormat = kong2kic.KICV3INGRESS
		case version == "2" && !ingress:
			outputFileFormat = kong2kic.KICV2GATEWAY
		case version == "2" && ingress:
			outputFileFormat = kong2kic.KICV2INGRESS
		default:
			err = fmt.Errorf("invalid combination of parameters. Please use --help for more information")
		}

		if outputFormat == file.YAML {
			yamlOrJSON = file.YAML
		} else if outputFormat == file.JSON {
			yamlOrJSON = file.JSON
		}
	}

	return outputFileFormat, yamlOrJSON, err
}

//
//
// Define the CLI data for the kong2kic command
//
//

func newKong2KicCmd() *cobra.Command {
	kong2KicCmd := &cobra.Command{
		Use:   "kong2kic",
		Short: "Convert Kong configuration files to Kong Ingress Controller (KIC) manifests",
		Long: `Convert Kong configuration files to Kong Ingress Controller (KIC) manifests.
		
The kong2kic subcommand transforms Kong’s configuration files, written in the deck format, 
into Kubernetes manifests suitable for the Kong Ingress Controller. By default kong2kic generates 
manifests for KIC v3.x using the Kubernetes Gateway API. Only HTTP/HTTPS routes are supported.`,
		RunE: executeKong2Kic,
		Args: cobra.NoArgs,
	}

	kong2KicCmd.Flags().StringVarP(&cmdKong2KicInputFilename, "state", "s", "-",
		"decK file to process. Use - to read from stdin.")
	kong2KicCmd.Flags().StringVarP(&cmdKong2KicOutputFilename, "output-file", "o", "-",
		"Output file to write. Use - to write to stdout.")
	kong2KicCmd.Flags().StringVarP(&cmdKong2KicOutputFormat, "format", "f", "yaml",
		"Output file format: json or yaml.")
	kong2KicCmd.Flags().StringVar(&cmdKong2KicVersion, "kic-version", "3",
		"Generate manifests for KIC v3 or v2. Possible values are 2 or 3.")
	kong2KicCmd.Flags().StringVar(&cmdKong2KicClassName, "class-name", "kong",
		`Value to use for "kubernetes.io/ingress.class" ObjectMeta.Annotations and for
		"parentRefs.name" in the case of HTTPRoute.`)
	kong2KicCmd.Flags().BoolVar(&cmdKong2KicIngress, "ingress", false,
		`Use Kubernetes Ingress API manifests instead of Gateway API manifests.`)

	return kong2KicCmd
}
