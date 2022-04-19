package cmd

import (
	"fmt"

	"github.com/kong/deck/convert"
	"github.com/kong/deck/cprint"
	"github.com/kong/deck/utils"
	"github.com/spf13/cobra"
)

var (
	convertCmdSourceFormat      string
	convertCmdDestinationFormat string
	convertCmdInputFile         string
	convertCmdOutputFile        string
)

// newConvertCmd represents the convert command
func newConvertCmd() *cobra.Command {
	convertCmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert files from one format into another format",
		Long: `The convert command changes configuration files from one format
into another compatible format. For example, a configuration for 'kong-gateway'
can be converted into a 'konnect' configuration file.`,
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
			if convertCmdDestinationFormat == "konnect" {
				cprint.UpdatePrintf("Warning: konnect format type was deprecated in v1.12 and it will be removed\n" +
					"in a future version. Please use your Kong configuration files with deck <cmd>.\n" +
					"Please see https://docs.konghq.com/konnect/deployment/import.\n")
			}
			return nil
		},
	}

	sourceFormats := []convert.Format{convert.FormatKongGateway}
	destinationFormats := []convert.Format{convert.FormatKonnect}
	convertCmd.Flags().StringVar(&convertCmdSourceFormat, "from", "",
		fmt.Sprintf("format of the source file, allowed formats: %v", sourceFormats))
	convertCmd.Flags().StringVar(&convertCmdDestinationFormat, "to", "",
		fmt.Sprintf("desired format of the output, allowed formats: %v", destinationFormats))
	convertCmd.Flags().StringVar(&convertCmdInputFile, "input-file", "",
		"configuration file to be converted. Use `-` to read from stdin.")
	convertCmd.Flags().StringVar(&convertCmdOutputFile, "output-file", "",
		"file to write configuration to after conversion. Use `-` to write to stdout.")
	return convertCmd
}
