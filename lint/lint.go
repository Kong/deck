package lint

import (
	"fmt"
	"strings"

	"github.com/daveshanley/vacuum/motor"
	"github.com/daveshanley/vacuum/rulesets"
	"github.com/kong/go-apiops/filebasics"
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

type Result struct {
	Message   string
	Severity  string
	Line      int
	Column    int
	Character int
	Path      string
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

func Lint(
	cmdLintInputFilename string,
	ruleSetFileName string,
	cmdLintFailSeverity string,
	cmdLintOnlyFailures bool,
) (map[string]interface{}, error) {
	customRuleSet, err := getRuleSet(ruleSetFileName)
	if err != nil {
		return nil, err
	}

	stateFileBytes, err := filebasics.ReadFile(cmdLintInputFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file '%s'; %w", cmdLintInputFilename, err)
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
		lintResults  = make([]Result, 0)
	)
	for _, x := range ruleSetResults.Results {
		if cmdLintOnlyFailures && ParseSeverity(x.Rule.Severity) < ParseSeverity(cmdLintFailSeverity) {
			continue
		}
		if ParseSeverity(x.Rule.Severity) >= ParseSeverity(cmdLintFailSeverity) {
			failingCount++
		}
		totalCount++
		lintResults = append(lintResults, Result{
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

	return lintErrs, nil
}

func GetLintOutput(lintErrs map[string]interface{}, cmdLintFormat, cmdLintOutputFilename string) (bool, error) {
	outputFormat := strings.ToUpper(cmdLintFormat)
	totalCount := lintErrs["total_count"].(int)
	failingCount := lintErrs["fail_count"].(int)
	silenceErrors := false

	switch outputFormat {
	case strings.ToUpper(string(filebasics.OutputFormatJSON)):
		fallthrough
	case strings.ToUpper(string(filebasics.OutputFormatYaml)):
		if err := filebasics.WriteSerializedFile(
			cmdLintOutputFilename, lintErrs, filebasics.OutputFormat(outputFormat),
		); err != nil {
			return silenceErrors, fmt.Errorf("error writing lint results: %w", err)
		}
	case strings.ToUpper(plainTextFormat):
		if totalCount > 0 {
			fmt.Printf("Linting Violations: %d\n", totalCount)
			fmt.Printf("Failures: %d\n\n", failingCount)
			for _, violation := range lintErrs["results"].([]Result) {
				fmt.Printf("[%s][%d:%d] %s\n",
					violation.Severity, violation.Line, violation.Column, violation.Message,
				)
			}
		}
	default:
		return silenceErrors, fmt.Errorf("invalid output format: %s", cmdLintFormat)
	}

	if failingCount > 0 {
		silenceErrors = true
	}

	return silenceErrors, nil
}
