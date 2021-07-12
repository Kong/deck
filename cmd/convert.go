package cmd

import (
	"fmt"

	"github.com/kong/deck/convert"
	"github.com/kong/deck/utils"
	"github.com/spf13/cobra"
)

var (
	convertCmdSourceFormat      string
	convertCmdDestinationFormat string
	convertCmdInputFile         string
	convertCmdOutputFile        string
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert files in one format to another format",
	Long: `Convert command converts files representing configuration in one format
to another compatible format. For example: a configuration for 'kong-gateway'
can be converted into 'konnect' configuration file.`,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceFormat, err := convert.ParseFormat(convertCmdSourceFormat)
		if err != nil {
			return err
		}
		destinationFormat, err := convert.ParseFormat(convertCmdDestinationFormat)
		if err != nil {
			return err
		}

		if yes, err := utils.ConfirmFileOverwrite(convertCmdOutputFile, "", false); err != nil {
			return err
		} else if !yes {
			return nil
		}

		err = convert.Convert(convertCmdInputFile, convertCmdOutputFile, sourceFormat, destinationFormat)
		if err != nil {
			return fmt.Errorf("converting file: %v", err)
		}
		return nil
	},
}

func init() {
	sourceFormats := []convert.Format{convert.FormatKongGateway}
	destinationFormats := []convert.Format{convert.FormatKonnect}
	convertCmd.Flags().StringVar(&convertCmdSourceFormat, "from", "",
		fmt.Sprintf("format of the source file, allowed formats: %v", sourceFormats))
	convertCmd.Flags().StringVar(&convertCmdDestinationFormat, "to", "",
		fmt.Sprintf("desired format of the output, allowed formats: %v", destinationFormats))
	convertCmd.Flags().StringVar(&convertCmdInputFile, "input-file", "",
		"file containing configuration that needs to be converted. Use '-' to read from stdin.")
	convertCmd.Flags().StringVar(&convertCmdOutputFile, "output-file", "",
		"file to which to write configuration after conversion. Use '-' to write to stdout.")
	rootCmd.AddCommand(convertCmd)
}
