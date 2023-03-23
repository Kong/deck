package cmd

import (
	"fmt"
	"os"

	"github.com/kong/deck/convert"
	"github.com/kong/deck/cprint"
	"github.com/kong/deck/utils"
	"github.com/spf13/cobra"
)

var (
	convertCmdSourceFormat       string
	convertCmdDestinationFormat  string
	convertCmdInputFile          []string
	convertCmdOutputFile         string
	convertCmdAssumeYes          bool
	convertCmdDisableMockEnvVars bool
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

			if len(convertCmdInputFile) != 0 {
				if yes, err := utils.ConfirmFileOverwrite(
					convertCmdOutputFile, "", convertCmdAssumeYes,
				); err != nil {
					return err
				} else if !yes {
					return nil
				}

				err = convert.Convert(convertCmdInputFile,
					convertCmdOutputFile,
					sourceFormat,
					destinationFormat,
					!convertCmdDisableMockEnvVars)
				if err != nil {
					return fmt.Errorf("converting file: %v", err)
				}
			} else if is2xTo3xConversion() {
				path, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("getting current working directory: %w", err)
				}
				files, err := utils.ConfigFilesInDir(path)
				if err != nil {
					return fmt.Errorf("getting files from directory: %w", err)
				}
				for _, filename := range files {
					err = convert.Convert([]string{filename}, filename, sourceFormat, destinationFormat, !convertCmdDisableMockEnvVars)
					if err != nil {
						return fmt.Errorf("converting '%s' file: %v", filename, err)
					}
				}
			}
			if convertCmdDestinationFormat == "konnect" {
				cprint.UpdatePrintf("Warning: konnect format type was deprecated in v1.12 and it will be removed\n" +
					"in a future version. Please use your Kong configuration files with deck <cmd>.\n" +
					"Please see https://docs.konghq.com/konnect/deployment/import.\n")
			}
			return nil
		},
	}

	sourceFormats := []convert.Format{convert.FormatKongGateway, convert.FormatKongGateway2x, convert.FormatDistributed}
	destinationFormats := []convert.Format{convert.FormatKonnect, convert.FormatKongGateway3x}
	convertCmd.Flags().StringVar(&convertCmdSourceFormat, "from", "",
		fmt.Sprintf("format of the source file, allowed formats: %v", sourceFormats))
	convertCmd.Flags().StringVar(&convertCmdDestinationFormat, "to", "",
		fmt.Sprintf("desired format of the output, allowed formats: %v", destinationFormats))
	convertCmd.Flags().StringSliceVar(&convertCmdInputFile, "input-file", []string{},
		"configuration files to be converted. Use `-` to read from stdin.")
	convertCmd.Flags().StringVar(&convertCmdOutputFile, "output-file", "kong.yaml",
		"file to write configuration to after conversion. Use `-` to write to stdout.")
	convertCmd.Flags().BoolVar(&convertCmdAssumeYes, "yes",
		false, "assume `yes` to prompts and run non-interactively.")
	convertCmd.Flags().BoolVar(&convertCmdDisableMockEnvVars, "disable-mock-env",
		false, "disables the mocking of environment variables.")
	return convertCmd
}

func is2xTo3xConversion() bool {
	return convertCmdSourceFormat == string(convert.FormatKongGateway2x) &&
		convertCmdDestinationFormat == string(convert.FormatKongGateway3x)
}
