package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kong/deck/convert"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	convertCmdSourceFormat         string
	convertCmdDestinationFormat    string // konnect/kong-gateway-3.x/etc
	convertCmdInputFile            string
	convertCmdOutputFile           string
	convertCmdInputFileDeprecated  string
	convertCmdOutputFileDeprecated string
	convertCmdAssumeYes            bool
	convertCmdStateFormat          string // yaml/json output
)

func executeConvert(_ *cobra.Command, _ []string) error {
	_ = sendAnalytics("file-convert", "", modeLocal)

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

		err = convert.Convert(
			[]string{convertCmdInputFile},
			convertCmdOutputFile,
			file.Format(strings.ToUpper(convertCmdStateFormat)),
			sourceFormat,
			destinationFormat,
			false)
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
			err = convert.Convert(
				[]string{filename},
				filename,
				file.Format(strings.ToUpper(convertCmdStateFormat)),
				sourceFormat,
				destinationFormat,
				false)
			if err != nil {
				return fmt.Errorf("converting '%s' file: %w", filename, err)
			}
		}
	}
	if convertCmdDestinationFormat == "konnect" {
		fmt.Fprintf(os.Stderr, "Warning: konnect format type was deprecated in v1.12 and it will be removed\n"+
			"in a future version. Please use your Kong configuration files with deck <cmd>.\n"+
			"Please see https://docs.konghq.com/konnect/getting-started/import/.\n")
	}
	return nil
}

// newConvertCmd represents the convert command
func newConvertCmd(deprecated bool) *cobra.Command {
	short := "Convert files from one format into another format"
	execute := executeConvert
	fileInDefault := "-"
	fileOutDefault := "-"
	if deprecated {
		short = "[deprecated] see 'deck file convert --help' for changes to the command"
		execute = func(cmd *cobra.Command, args []string) error {
			convertCmdInputFile = convertCmdInputFileDeprecated
			convertCmdOutputFile = convertCmdOutputFileDeprecated
			fmt.Fprintf(os.Stderr, "Info: 'deck convert' functionality has moved to 'deck file convert' and will be removed\n"+
				"in a future MAJOR version of deck. Migration to 'deck file convert' is recommended.\n"+
				"   Note: - see 'deck file convert --help' for changes to the command\n"+
				"         - files changed to positional arguments without the '-s/--state' flag\n"+
				"         - the default changed from 'kong.yaml' to '-' (stdin/stdout)\n")
			return executeConvert(cmd, args)
		}
		fileInDefault = ""
		fileOutDefault = "kong.yaml"
	}

	convertCmd := &cobra.Command{
		Use:   "convert",
		Short: short,
		Long: `The convert command changes configuration files from one format
into another compatible format. For example, a configuration for 'kong-gateway-2.x'
can be converted into a 'kong-gateway-3.x' configuration file.`,
		Args: validateNoArgs,
		RunE: execute,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			validSourceFormats := []string{string(convert.FormatKongGateway), string(convert.FormatKongGateway2x)}
			validDestinationFormats := []string{string(convert.FormatKonnect), string(convert.FormatKongGateway3x)}

			err := validateInputFlag("from", convertCmdSourceFormat, validSourceFormats, "")
			if err != nil {
				return err
			}

			err = validateInputFlag("to", convertCmdDestinationFormat, validDestinationFormats, "")
			if err != nil {
				return err
			}

			return nil
		},
	}

	sourceFormats := []convert.Format{convert.FormatKongGateway, convert.FormatKongGateway2x}
	destinationFormats := []convert.Format{convert.FormatKonnect, convert.FormatKongGateway3x}
	convertCmd.Flags().StringVar(&convertCmdSourceFormat, "from", "",
		fmt.Sprintf("format of the source file, allowed formats: %v", sourceFormats))
	convertCmd.Flags().StringVar(&convertCmdDestinationFormat, "to", "",
		fmt.Sprintf("desired format of the output, allowed formats: %v", destinationFormats))
	if deprecated {
		convertCmd.Flags().StringVar(&convertCmdInputFileDeprecated, "input-file", fileInDefault,
			"configuration file to be converted. Use `-` to read from stdin.")
		convertCmd.Flags().StringVarP(&convertCmdOutputFileDeprecated, "output-file", "o", fileOutDefault,
			"file to write configuration to after conversion. Use `-` to write to stdout.")
	} else {
		convertCmd.Flags().StringVar(&convertCmdInputFile, "input-file", fileInDefault,
			"configuration file to be converted. Use `-` to read from stdin.")
		convertCmd.Flags().StringVarP(&convertCmdOutputFile, "output-file", "o", fileOutDefault,
			"file to write configuration to after conversion. Use `-` to write to stdout.")
	}
	convertCmd.Flags().BoolVar(&convertCmdAssumeYes, "yes",
		false, "assume `yes` to prompts and run non-interactively.")
	convertCmd.Flags().StringVar(&convertCmdStateFormat, "format",
		"yaml", "output file format: json or yaml.")

	return convertCmd
}

func is2xTo3xConversion() bool {
	return convertCmdSourceFormat == string(convert.FormatKongGateway2x) &&
		convertCmdDestinationFormat == string(convert.FormatKongGateway3x)
}
