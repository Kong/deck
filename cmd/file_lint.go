package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/daveshanley/vacuum/motor"
	"github.com/daveshanley/vacuum/rulesets"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/spf13/cobra"
)

var (
	cmdLintInputFilename string
	cmdLintInputRuleset  string
	cmdLintFormat        string
	cmdLintFailSeverity  string
	cmdLintOnlyFailures  bool
)

const plainTextFormat = "plain"

type Severity int

const (
	SeverityHint Severity = iota
	SeverityInfo
	SeverityWarn
	SeverityError
)

var severityStrings = [...]string{
	"hint",
	"info",
	"warn",
	"error",
}

type LintResult struct {
	Message   string
	Severity  string
	Line      int
	Column    int
	Character int
	Path      string
}

func ParseSeverity(s string) Severity {
	for i, str := range severityStrings {
		if s == str {
			return Severity(i)
		}
	}
	return SeverityWarn
}

func getRuleSet(ruleSetFile string) (*rulesets.RuleSet, error) {
	ruleSetBytes, err := os.ReadFile(ruleSetFile)
	if err != nil {
		return nil, fmt.Errorf("error reading ruleset file: %w", err)
	}
	customRuleSet, err := rulesets.CreateRuleSetFromData(ruleSetBytes)
	if err != nil {
		return nil, fmt.Errorf("error creating ruleset: %w", err)
	}
	return customRuleSet, nil
}

func executeLint(cmd *cobra.Command, _ []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)

	if cmdLintInputRuleset == "" {
		return errors.New("missing required option: --ruleset")
	}
	customRuleSet, err := getRuleSet(cmdLintInputRuleset)
	if err != nil {
		return err
	}

	var stateFileBytes []byte
	if cmdLintInputFilename == "-" {
		stateFileBytes, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading state from STDIN: %w", err)
		}
	} else {
		stateFileBytes, err = os.ReadFile(cmdLintInputFilename)
		if err != nil {
			return fmt.Errorf("error reading state file: %w", err)
		}
	}

	ruleSetResults := motor.ApplyRulesToRuleSet(&motor.RuleSetExecution{
		RuleSet:           customRuleSet,
		Spec:              stateFileBytes,
		SkipDocumentCheck: true,
	})

	var (
		failingCount int
		totalCount   int
		lintResults  = make([]LintResult, 0)
	)
	for _, x := range ruleSetResults.Results {
		if cmdLintOnlyFailures && ParseSeverity(x.Rule.Severity) < ParseSeverity(cmdLintFailSeverity) {
			continue
		}
		if ParseSeverity(x.Rule.Severity) >= ParseSeverity(cmdLintFailSeverity) {
			failingCount++
		}
		totalCount++
		lintResults = append(lintResults, LintResult{
			Message: x.Message,
			Path: func() string {
				if path, ok := x.Rule.Given.(string); ok {
					return path
				}
				return ""
			}(),
			Line:     x.StartNode.Line,
			Column:   x.StartNode.Column,
			Severity: x.Rule.Severity,
		})
	}

	lintErrs := map[string]interface{}{
		"total_count": totalCount,
		"fail_count":  failingCount,
		"results":     lintResults,
	}

	outputFormat := strings.ToUpper(cmdLintFormat)
	switch outputFormat {
	case strings.ToUpper(string(filebasics.OutputFormatJSON)):
		fallthrough
	case strings.ToUpper(string(filebasics.OutputFormatYaml)):
		if err = filebasics.WriteSerializedFile(
			"-", lintErrs, filebasics.OutputFormat(outputFormat),
		); err != nil {
			return fmt.Errorf("error writing lint results: %w", err)
		}
	case strings.ToUpper(plainTextFormat):
		if totalCount > 0 {
			fmt.Printf("Linting Violations: %d\n", totalCount)
			fmt.Printf("Failures: %d\n\n", failingCount)
			for _, violation := range lintErrs["results"].([]LintResult) {
				fmt.Printf("[%s][%d:%d] %s\n",
					violation.Severity, violation.Line, violation.Column, violation.Message,
				)
			}
		}
	default:
		return fmt.Errorf("invalid output format: %s", cmdLintFormat)
	}
	if failingCount > 0 {
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
		Use:   "lint",
		Short: "Lint a file against a ruleset",
		Long: "Validate a decK state file against a linting ruleset, reporting any violations or failures.\n" +
			"Report output can be returned in JSON, YAML, or human readable format (see --format).\n" +
			"Ruleset Docs: https://quobix.com/vacuum/rulesets/",
		RunE: executeLint,
	}

	lintCmd.Flags().StringVarP(&cmdLintInputFilename, "state", "s", "-",
		"decK file to process. Use - to read from stdin.")
	lintCmd.Flags().StringVarP(&cmdLintInputRuleset, "ruleset", "r", "",
		"Ruleset to apply to the state file.")
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
		"results of this level or above will trigger a failure exit code "+
			"[choices: \"error\", \"warn\", \"info\", \"hint\"]")
	lintCmd.Flags().BoolVarP(&cmdLintOnlyFailures,
		"display-only-failures", "D", false,
		"only output results equal to or greater than --fail-severity")

	return lintCmd
}
