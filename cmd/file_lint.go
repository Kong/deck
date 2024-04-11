package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/daveshanley/vacuum/motor"
	"github.com/daveshanley/vacuum/rulesets"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var (
	cmdLintInputFilename  string
	cmdLintFormat         string
	cmdLintFailSeverity   string
	cmdLintOutputFilename string
	cmdLintOnlyFailures   bool
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

func isOpenAPISpec(fileBytes []byte) bool {
	var contents map[string]interface{}

	// This marshalling is redundant with what happens
	// in the linting command. There is likely an algorithm
	// we could use to determine JSON vs YAML and pull out the
	// openapi key without unmarshalling the entire file.
	err := json.Unmarshal(fileBytes, &contents)
	if err != nil {
		err = yaml.Unmarshal(fileBytes, &contents)
	}

	if err != nil {
		return false
	}

	return contents["openapi"] != nil
}

// getRuleSet reads the ruleset file by the provided name and returns a RuleSet object.
func getRuleSet(ruleSetFile string) (*rulesets.RuleSet, error) {
	ruleSetBytes, err := filebasics.ReadFile(ruleSetFile)
	if err != nil {
		return nil, fmt.Errorf("error reading ruleset file: %w", err)
	}

	customRuleSet, err := rulesets.CreateRuleSetFromData(ruleSetBytes)
	if err != nil {
		return nil, fmt.Errorf("error creating ruleset: %w", err)
	}

	extends := customRuleSet.GetExtendsValue()
	if len(extends) > 0 {
		defaultRuleSet := rulesets.BuildDefaultRuleSets()
		return defaultRuleSet.GenerateRuleSetFromSuppliedRuleSet(customRuleSet), nil
	}

	return customRuleSet, nil
}

// Executes the CLI command "lint"
func executeLint(cmd *cobra.Command, args []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)
	_ = sendAnalytics("file-lint", "", modeLocal)

	customRuleSet, err := getRuleSet(args[0])
	if err != nil {
		return err
	}

	stateFileBytes, err := filebasics.ReadFile(cmdLintInputFilename)
	if err != nil {
		return fmt.Errorf("failed to read input file '%s'; %w", cmdLintInputFilename, err)
	}

	ruleSetResults := motor.ApplyRulesToRuleSet(&motor.RuleSetExecution{
		RuleSet:           customRuleSet,
		Spec:              stateFileBytes,
		SkipDocumentCheck: !isOpenAPISpec(stateFileBytes),
		AllowLookup:       true,
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
			cmdLintOutputFilename, lintErrs, filebasics.OutputFormat(outputFormat),
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
