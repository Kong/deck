package cmd

import (
	"errors"
	"slices"

	"github.com/fatih/color"
	"github.com/kong/go-database-reconciler/pkg/cprint"
	"github.com/kong/go-database-reconciler/pkg/diff"
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

func validateStringOneOf(value string, allowedValues []string, errorMessage string) error {
	if slices.Contains(allowedValues, value) {
		return nil
	}

	return errors.New(errorMessage)
}
