package cmd

import (
	"fmt"
	"log"

	"github.com/kong/deck/kong2tf"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/spf13/cobra"
)

var (
	cmdKong2TfInputFilename  string
	cmdKong2TfOutputFilename string
	cmdKong2TfControlPlaneId string
)

// Executes the CLI command "kong2Tf"
func executeKong2Tf(cmd *cobra.Command, _ []string) error {
	_ = sendAnalytics("file-kong2Tf", "", modeLocal)
	var (
		result string
		err    error
	)

	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)

	logbasics.Info("Starting execution of executeKong2Tf")

	inputContent, err := file.GetContentFromFiles([]string{cmdKong2TfInputFilename}, false)
	if err != nil {
		log.Printf("Error reading input file '%s'; %v", cmdKong2TfInputFilename, err)
		return fmt.Errorf("failed reading input file '%s'; %w", cmdKong2TfInputFilename, err)
	}
	logbasics.Info("Successfully read input file '%s'", cmdKong2TfInputFilename)

	logbasics.Info("Converting Kong configuration to Terraform")
	result, err = kong2tf.Convert(inputContent)
	if err != nil {
		log.Printf("Error converting Kong configuration to Terraform; %v", err)
		return fmt.Errorf("failed converting Kong configuration to Terraform; %w", err)
	}
	logbasics.Info("Successfully converted Kong configuration to Terraform")

	logbasics.Info("Writing output to file '%s'", cmdKong2TfOutputFilename)
	err = filebasics.WriteFile(cmdKong2TfOutputFilename, []byte(result))
	if err != nil {
		log.Printf("Error writing output to file '%s'; %v", cmdKong2TfOutputFilename, err)
		return err
	}
	logbasics.Info("Successfully wrote output to file '%s'", cmdKong2TfOutputFilename)

	logbasics.Info("Finished execution of executeKong2Tf")
	return nil
}

//
//
// Define the CLI data for the kong2Tf command
//
//

func newKong2TfCmd() *cobra.Command {
	kong2TfCmd := &cobra.Command{
		Use:   "kong2tf",
		Short: "Convert Kong configuration files to Terraform resources",
		Long: `Convert Kong configuration files to Terraform resources.
		
The kong2tf subcommand transforms Kong Gateway entities in deck format, 
into Terraform resources.`,
		RunE: executeKong2Tf,
		Args: cobra.NoArgs,
	}

	kong2TfCmd.Flags().StringVarP(&cmdKong2TfInputFilename, "state", "s", "-",
		"decK file to process. Use - to read from stdin.")
	kong2TfCmd.Flags().StringVarP(&cmdKong2TfOutputFilename, "output-file", "o", "-",
		"Output file to write. Use - to write to stdout.")

	return kong2TfCmd
}
