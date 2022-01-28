package cmd

import (
	"github.com/fatih/color"
	"github.com/kong/deck/cprint"
	"github.com/kong/deck/diff"
	"github.com/spf13/pflag"
)

func printStats(stats diff.Stats, workspace string) {
	// do not use github.com/kong/deck/print because that package
	// is used only for events logs
	printFn := color.New(color.FgGreen, color.Bold).PrintfFunc()
	if workspace != "" {
		printFn("Summary (%s):\n", workspace)
	} else {
		printFn("Summary:\n")
	}
	printFn("  Created: %d\n", stats.CreateOps.Count())
	printFn("  Updated: %d\n", stats.UpdateOps.Count())
	printFn("  Deleted: %d\n", stats.DeleteOps.Count())
	printFn("\n")
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
