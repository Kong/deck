package cmd

import (
	"errors"
	"fmt"
	"os"
	"slices"

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

// expandToFiles expands a list of file/directory paths into a flat list of
// individual files. Directories are walked recursively.
func expandToFiles(paths []string) ([]string, error) {
	var result []string
	for _, p := range paths {
		fi, err := os.Stat(p)
		if err != nil || !fi.IsDir() {
			result = append(result, p)
			continue
		}
		files, err := reconcilerUtils.ConfigFilesInDir(p)
		if err != nil {
			return nil, err
		}
		result = append(result, files...)
	}
	return result, nil
}

// batchFiles splits files into consecutive batches of at most size entries.
// If size < 1, each file is its own batch (equivalent to size=1).
func batchFiles(files []string, size int) [][]string {
	if size < 1 {
		size = 1
	}
	var batches [][]string
	for i := 0; i < len(files); i += size {
		end := i + size
		if end > len(files) {
			end = len(files)
		}
		batches = append(batches, files[i:end])
	}
	return batches
}
