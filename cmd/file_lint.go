package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/kong/deck/lint"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/spf13/cobra"
)

var (
	cmdLintInputFilename  string
	cmdLintFormat         string
	cmdLintFailSeverity   string
	cmdLintOutputFilename string
	cmdLintOnlyFailures   bool
)

const plainTextFormat = "plain"

// Executes the CLI command "lint"
func executeLint(cmd *cobra.Command, args []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)
	_ = sendAnalytics("file-lint", "", modeLocal)

	lintErrs, err := lint.Lint(cmdLintInputFilename, args[0],
		cmdLintFailSeverity, cmdLintOnlyFailures)
	if err != nil {
		return err
	}

	silenceErrors, err := lint.GetLintOutput(lintErrs, cmdLintFormat, cmdLintOutputFilename)
	if err != nil {
		return err
	}

	if silenceErrors {
		// We don't want to print the error here as they're already output above
		// But we _do_ want to set an exit code of failure.
		//
		// We could simply use os.Exit(1) here, but that would make e2e tests harder.
		cmd.SilenceErrors = true
		return errors.New("linting errors detected")
	}
	return nil
}

//
//
// Define the CLI data for the lint command
//
//

func newLintCmd() *cobra.Command {
	lintCmd := &cobra.Command{
		Use:   "lint [flags] ruleset-file",
		Short: "Lint a file against a ruleset",
		Long: "Validate a decK state file against a linting ruleset, reporting any violations or failures.\n" +
			"Report output can be returned in JSON, YAML, or human readable format (see --format).\n" +
			"Ruleset Docs: https://quobix.com/vacuum/rulesets/",
		RunE: executeLint,
		Args: cobra.ExactArgs(1),
	}

	lintCmd.Flags().StringVarP(&cmdLintInputFilename, "state", "s", "-",
		"decK file to process. Use - to read from stdin.")
	lintCmd.Flags().StringVarP(&cmdLintOutputFilename, "output-file", "o", "-",
		"Output file to write to. Use - to write to stdout.")
	lintCmd.Flags().StringVar(
		&cmdLintFormat, "format", plainTextFormat,
		fmt.Sprintf(`output format [choices: "%s", "%s", "%s"]`,
			plainTextFormat,
			string(filebasics.OutputFormatJSON),
			string(filebasics.OutputFormatYaml),
		),
	)
	lintCmd.Flags().StringVarP(
		&cmdLintFailSeverity, "fail-severity", "F", "error",
		"results of this level or above will trigger a failure exit code\n"+
			"[choices: \"error\", \"warn\", \"info\", \"hint\"]")
	lintCmd.Flags().BoolVarP(&cmdLintOnlyFailures,
		"display-only-failures", "D", false,
		"only output results equal to or greater than --fail-severity")

	return lintCmd
}
