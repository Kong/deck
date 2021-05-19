package cmd

import (
	"github.com/fatih/color"
	printPkg "github.com/kong/deck/print"
	"github.com/kong/deck/solver"
	"github.com/spf13/pflag"
)

func printStats(stats solver.Stats) {
	// do not use github.com/kong/deck/print because that package
	// is used only for events logs
	printFn := color.New(color.FgGreen, color.Bold).PrintfFunc()
	printFn("Summary:\n")
	printFn("  Created: %v\n", stats.CreateOps)
	printFn("  Updated: %v\n", stats.UpdateOps)
	printFn("  Deleted: %v\n", stats.DeleteOps)
}

var (
	silenceEvents bool
)

func preRunSilenceEventsFlag() error {
	printPkg.DisableOutput = true
	return nil
}

func addSilenceEventsFlag(set *pflag.FlagSet) {
	set.BoolVar(&silenceEvents, "silence-events", false,
		"disable printing events to stdout")
}
