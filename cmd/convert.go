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
	convertCmdSourceFormat      string
	convertCmdDestinationFormat string
	convertCmdInputFile         string
	convertCmdOutputFile        string
	convertCmdAssumeYes         bool
)

func executeConvert(_ *cobra.Command, _ []string) error {
	sourceFormat, err := convert.ParseFormat(convertCmdSourceFormat)
	if err != nil {
		return err
	}
	destinationFormat, err := convert.ParseFormat(convertCmdDestinationFormat)
	if err != nil {
		return err
	}

	if convertCmdInputFile != "" {
		if yes, err := utils.ConfirmFileOverwrite(
			convertCmdOutputFile, "", convertCmdAssumeYes,
		); err != nil {
			return err
		} else if !yes {
			return nil
		}

		err = convert.Convert(convertCmdInputFile, convertCmdOutputFile, sourceFormat, destinationFormat)
		if err != nil {
			return fmt.Errorf("converting file: %w", err)
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
			err = convert.Convert(filename, filename, sourceFormat, destinationFormat)
			if err != nil {
				return fmt.Errorf("converting '%s' file: %w", filename, err)
			}
		}
	}
	if convertCmdDestinationFormat == "konnect" {
		cprint.UpdatePrintf("Warning: konnect format type was deprecated in v1.12 and it will be removed\n" +
			"in a future version. Please use your Kong configuration files with deck <cmd>.\n" +
			"Please see https://docs.konghq.com/konnect/getting-started/import/.\n")
	}
	return nil
}

// newConvertCmd represents the convert command
func newConvertCmd() *cobra.Command {
	short := "Convert files from one format into another format"
	execute := executeConvert

	convertCmd := &cobra.Command{
		Use:   "convert",
		Short: short,
		Long: `The convert command changes configuration files from one format
into another compatible format. For example, a configuration for 'kong-gateway-2.x'
can be converted into a 'kong-gateway-3.x' configuration file.`,
		Args: validateNoArgs,
		RunE: execute,
	}

	sourceFormats := []convert.Format{convert.FormatKongGateway, convert.FormatKongGateway2x}
	destinationFormats := []convert.Format{convert.FormatKonnect, convert.FormatKongGateway3x}
	convertCmd.Flags().StringVar(&convertCmdSourceFormat, "from", "",
		fmt.Sprintf("format of the source file, allowed formats: %v", sourceFormats))
	convertCmd.Flags().StringVar(&convertCmdDestinationFormat, "to", "",
		fmt.Sprintf("desired format of the output, allowed formats: %v", destinationFormats))
	convertCmd.Flags().StringVar(&convertCmdInputFile, "input-file", "",
		"configuration file to be converted. Use `-` to read from stdin.")
	convertCmd.Flags().StringVar(&convertCmdOutputFile, "output-file", "kong.yaml",
		"file to write configuration to after conversion. Use `-` to write to stdout.")
	convertCmd.Flags().BoolVar(&convertCmdAssumeYes, "yes",
		false, "assume `yes` to prompts and run non-interactively.")
	return convertCmd
}

func is2xTo3xConversion() bool {
	return convertCmdSourceFormat == string(convert.FormatKongGateway2x) &&
		convertCmdDestinationFormat == string(convert.FormatKongGateway3x)
}
