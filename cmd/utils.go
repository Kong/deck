package cmd

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/kong/go-database-reconciler/pkg/cprint"
	"github.com/kong/go-database-reconciler/pkg/diff"
	reconcilerUtils "github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/spf13/pflag"
)

func printStats(stats diff.Stats) {
	// do not use github.com/kong/deck/print because that package
	// is used only for events logs
	printFn := color.New(color.FgGreen, color.Bold).PrintfFunc()
	printFn("Summary:\n")
	printFn("  Created: %d\n", stats.CreateOps.Count())
	printFn("  Updated: %d\n", stats.UpdateOps.Count())
	printFn("  Deleted: %d\n", stats.DeleteOps.Count())
}

var silenceEvents bool

var (
	diagnosticAlwaysError   string
	diagnosticAlwaysWarning string
	diagnosticPolicy        reconcilerUtils.DiagnosticPolicy
)

func preRunSilenceEventsFlag() error {
	if silenceEvents {
		cprint.DisableOutput = true
	}
	return nil
}

func addSilenceEventsFlag(set *pflag.FlagSet) {
	set.BoolVar(&silenceEvents, "silence-events", false,
		"disable printing events to stdout")
}

func preRunDiagnosticPolicyFlags() error {
	policy, err := buildDiagnosticPolicy(diagnosticAlwaysError, diagnosticAlwaysWarning)
	if err != nil {
		return err
	}
	diagnosticPolicy = policy
	return nil
}

func addDiagnosticSeverityFlags(set *pflag.FlagSet) {
	set.StringVarP(&diagnosticAlwaysError, "warnings-as-errors", "E", "",
		"Treat the given comma-separated diagnostic codes as errors.")
	set.StringVarP(&diagnosticAlwaysWarning, "errors-as-warnings", "W", "",
		"Treat the given comma-separated diagnostic codes as warnings.")
}

func buildDiagnosticPolicy(alwaysErrorValue, alwaysWarningValue string) (reconcilerUtils.DiagnosticPolicy, error) {
	alwaysErrorCodes, err := parseDiagnosticCodesFlag("-E/--warnings-as-errors", alwaysErrorValue)
	if err != nil {
		return reconcilerUtils.DiagnosticPolicy{}, err
	}

	alwaysWarningCodes, err := parseDiagnosticCodesFlag("-W/--warnings-as-warnings", alwaysWarningValue)
	if err != nil {
		return reconcilerUtils.DiagnosticPolicy{}, err
	}

	return reconcilerUtils.NewDiagnosticPolicy(alwaysErrorCodes, alwaysWarningCodes), nil
}

func parseDiagnosticCodesFlag(flagName, value string) ([]reconcilerUtils.DiagnosticCode, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	codes, err := reconcilerUtils.ParseDiagnosticCodes(value)
	if err != nil {
		return nil, fmt.Errorf("invalid value for %s: %w. Valid diagnostic codes: %v",
			flagName, err, reconcilerUtils.ValidDiagnosticCodes())
	}
	return codes, nil
}

func validateInputFlag(flagName string, flagValue string, allowedValues []string, errorMessage string) error {
	if slices.Contains(allowedValues, flagValue) {
		return nil
	}

	if errorMessage != "" {
		return errors.New(errorMessage)
	}

	return fmt.Errorf("invalid value '%s' found for the '%s' flag. Allowed values: %v",
		flagValue, flagName, allowedValues)
}

func checkParallelism(parallelism int) error {
	if parallelism < 1 {
		return fmt.Errorf("--parallelism cannot be less than 1, got %d", parallelism)
	}
	return nil
}
