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
	cmdKong2KicKICv2          bool
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
	if strings.ToUpper(cmdKong2KicOutputFormat) == file.YAML && !cmdKong2KicKICv2 && !cmdKong2KicIngress {
		// default to KICv 3.x and Gateway API YAML
		outputFileFormat = kong2kic.KICV3GATEWAY
		yamlOrJSON = file.YAML
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == file.JSON && !cmdKong2KicKICv2 && !cmdKong2KicIngress {
		// KICv 3.x and Gateway API JSON
		outputFileFormat = kong2kic.KICV3GATEWAY
		yamlOrJSON = file.JSON
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == file.YAML && !cmdKong2KicKICv2 && cmdKong2KicIngress {
		// KICv 3.x and Ingress API YAML
		outputFileFormat = kong2kic.KICV3INGRESS
		yamlOrJSON = file.YAML
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == file.JSON && !cmdKong2KicKICv2 && cmdKong2KicIngress {
		// KICv 3.x and Ingress API JSON
		outputFileFormat = kong2kic.KICV3INGRESS
		yamlOrJSON = file.JSON
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == file.YAML && cmdKong2KicKICv2 && !cmdKong2KicIngress {
		// KICv 2.x and Gateway API YAML
		outputFileFormat = kong2kic.KICV2GATEWAY
		yamlOrJSON = file.YAML
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == file.JSON && cmdKong2KicKICv2 && !cmdKong2KicIngress {
		// KICv 2.x and Gateway API JSON
		outputFileFormat = kong2kic.KICV2GATEWAY
		yamlOrJSON = file.JSON
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == file.YAML && cmdKong2KicKICv2 && cmdKong2KicIngress {
		// KICv 2.x and Ingress API YAML
		outputFileFormat = kong2kic.KICV2INGRESS
		yamlOrJSON = file.YAML
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == file.JSON && cmdKong2KicKICv2 && cmdKong2KicIngress {
		// KICv 2.x and Ingress API JSON
		outputFileFormat = kong2kic.KICV2INGRESS
		yamlOrJSON = file.JSON
	} else {
		err = fmt.Errorf("invalid combination parameters. Please use --help for more information")
	}

	return outputFileFormat, yamlOrJSON, err
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
		"output file format: json or yaml.")
	kong2KicCmd.Flags().StringVar(&cmdKong2KicClassName, "class-name", "kong",
		`Value to use for "kubernetes.io/ingress.class" ObjectMeta.Annotations and for
		"parentRefs.name" in the case of HTTPRoute.`)
	kong2KicCmd.Flags().BoolVar(&cmdKong2KicIngress, "ingress", false,
		`Use Kubernetes Ingress API manifests instead of Gateway API manifests.`)
	kong2KicCmd.Flags().BoolVar(&cmdKong2KicKICv2, "kicv2", false,
		`Generate manifests compatible with KIC v2.x.`)

	return kong2KicCmd
}
