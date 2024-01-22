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
	cmdKong2KicAPI            string
	cmdKong2KicOutputFormat   string
	cmdKong2KicManifestStyle  string
	CmdKong2KicClassName      string
	CmdKong2KicV1beta1        bool
)

// Executes the CLI command "kong2kic"
func executeKong2Kic(cmd *cobra.Command, _ []string) error {
	_ = sendAnalytics("file-kong2kic", "", modeLocal)
	var (
		outputContent    *file.Content
		err              error
		outputFileFormat file.Format
	)

	if CmdKong2KicV1beta1 {
		kong2kic.GatewayAPIVersion = "gateway.networking.k8s.io/v1beta1"
	} else {
		kong2kic.GatewayAPIVersion = "gateway.networking.k8s.io/v1"
	}
	kong2kic.ClassName = CmdKong2KicClassName
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)

	inputContent, err := file.GetContentFromFiles([]string{cmdKong2KicInputFilename}, false)
	if err != nil {
		return fmt.Errorf("failed reading input file '%s'; %w", cmdKong2KicInputFilename, err)
	}

	outputFileFormat, err = validateInput(cmdKong2KicOutputFormat, cmdKong2KicAPI, cmdKong2KicManifestStyle)
	if err != nil {
		return err
	}

	outputContent = inputContent.DeepCopy()
	err = kong2kic.WriteContentToFile(outputContent, cmdKong2KicOutputFilename, outputFileFormat)

	if err != nil {
		return fmt.Errorf("failed converting Kong to Ingress '%s'; %w", cmdKong2KicInputFilename, err)
	}

	return nil
}

func validateInput(cmdKong2KicOutputFormat, cmdKong2KicAPI, cmdKong2KicManifestStyle string) (
	outputFileFormat file.Format, err error,
) {
	const (
		JSON       = "JSON"
		YAML       = "YAML"
		INGRESS    = "INGRESS"
		GATEWAY    = "GATEWAY"
		CRD        = "CRD"
		ANNOTATION = "ANNOTATION"
	)

	if strings.ToUpper(cmdKong2KicOutputFormat) == JSON &&
		strings.ToUpper(cmdKong2KicAPI) == INGRESS &&
		strings.ToUpper(cmdKong2KicManifestStyle) == CRD {
		outputFileFormat = kong2kic.KICJSONCrdIngressAPI
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == JSON &&
		strings.ToUpper(cmdKong2KicAPI) == INGRESS &&
		strings.ToUpper(cmdKong2KicManifestStyle) == ANNOTATION {
		outputFileFormat = kong2kic.KICJSONAnnotationIngressAPI
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == YAML &&
		strings.ToUpper(cmdKong2KicAPI) == INGRESS &&
		strings.ToUpper(cmdKong2KicManifestStyle) == CRD {
		outputFileFormat = kong2kic.KICYAMLCrdIngressAPI
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == YAML &&
		strings.ToUpper(cmdKong2KicAPI) == INGRESS &&
		strings.ToUpper(cmdKong2KicManifestStyle) == ANNOTATION {
		outputFileFormat = kong2kic.KICYAMLAnnotationIngressAPI
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == JSON &&
		strings.ToUpper(cmdKong2KicAPI) == GATEWAY {
		outputFileFormat = kong2kic.KICJSONGatewayAPI
	} else if strings.ToUpper(cmdKong2KicOutputFormat) == YAML &&
		strings.ToUpper(cmdKong2KicAPI) == GATEWAY {
		outputFileFormat = kong2kic.KICYAMLGatewayAPI
	} else {
		err = fmt.Errorf("invalid combination of output format and manifest style. Valid combinations are:\n" +
			"Ingress API with annotation style in YAML format: --format yaml --api ingress --style annotation\n" +
			"Ingress API with annotation style in JSON format: --format json --api ingress --style annotation\n" +
			"Ingress API with CRD style in YAML format: --format yaml --api ingress --style crd\n" +
			"Ingress API with CRD style in JSON format: --format json --api ingress --style crd\n" +
			"Gateway API in YAML format: --format yaml --api gateway\n" +
			"Gateway API in JSON format: --format json --api gateway")
	}

	return outputFileFormat, err
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
		
Manifests can be generated using the Ingress API or the Gateway API. Ingress API manifests 
can be generated using annotations in Ingress and Service objects (recommended) or
using KongIngress objects. Output in YAML or JSON format. Only HTTP/HTTPS routes are supported.`,
		RunE: executeKong2Kic,
		Args: cobra.NoArgs,
	}

	kong2KicCmd.Flags().StringVarP(&cmdKong2KicInputFilename, "state", "s", "-",
		"decK file to process. Use - to read from stdin.")
	kong2KicCmd.Flags().StringVarP(&cmdKong2KicOutputFilename, "output-file", "o", "-",
		"Output file to write. Use - to write to stdout.")
	kong2KicCmd.Flags().StringVarP(&cmdKong2KicOutputFormat, "format", "f", "yaml",
		"output file format: json or yaml.")
	kong2KicCmd.Flags().StringVar(&CmdKong2KicClassName, "class-name", "kong",
		`Value to use for "kubernetes.io/ingress.class" ObjectMeta.Annotations and for
		"parentRefs.name" in the case of HTTPRoute.`)
	kong2KicCmd.Flags().StringVarP(&cmdKong2KicAPI, "api", "a", "ingress",
		`Use Ingress API manifests or Gateway API manifests: ingress or gateway`)
	kong2KicCmd.Flags().BoolVar(&CmdKong2KicV1beta1, "v1beta1", false,
		`Only for Gateway API, setting this flag will use "apiVersion: gateway.networking.k8s.io/v1beta1"
		in Gateway API manifests. Otherwise, "apiVersion: gateway.konghq.com/v1" is used.
		KIC versions earlier than 3.0 only support v1beta1.`)
	kong2KicCmd.Flags().StringVar(&cmdKong2KicManifestStyle, "style", "annotation",
		`Only for Ingress API, generate manifests with annotations in Service objects 
		and Ingress objects, or use only KongIngress objects without annotations: annotation or crd.`)

	return kong2KicCmd
}
