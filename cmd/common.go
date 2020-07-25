package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/dump"
	"github.com/hbagdi/deck/file"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/solver"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

const (
	exitCodeDiffDetection = 2
)

var (
	stopChannel chan struct{}
	dumpConfig  dump.Config
)

// SetStopCh sets the stop channel for long running commands.
// This is useful for cases when a process needs to be cancelled gracefully
// before it can complete to finish. Example: SIGINT
func SetStopCh(stopCh chan struct{}) {
	stopChannel = stopCh
}

// workspaceExists checks if workspace exists in Kong.
func workspaceExists(config utils.KongClientConfig) (bool, error) {

	workspace := config.Workspace
	if workspace == "" {
		// default workspace always exists
		return true, nil
	}

	// remove workspace to be able to call top-level /workspaces endpoint
	config.Workspace = ""
	rootClient, err := utils.GetKongClient(config)
	if err != nil {
		return false, err
	}

	_, err = rootClient.Workspaces.Get(nil, &workspace)
	if err != nil {
		if kong.IsNotFoundErr(err) {
			return false, nil
		}

		return false, errors.Wrap(err, "error when getting workspace")
	}

	return true, nil
}

func syncMain(filenames []string, dry bool, parallelism, delay int) error {

	// load Kong version before workspace
	kongVersion, err := kongVersion(config)
	if err != nil {
		return errors.Wrap(err, "reading Kong version")
	}

	// read target file
	targetContent, err := file.GetContentFromFiles(filenames)
	if err != nil {
		return err
	}

	rootClient, err := utils.GetKongClient(config)
	if err != nil {
		return err
	}

	// prepare to read the current state from Kong
	config.Workspace = targetContent.Workspace

	workspaceExists, err := workspaceExists(config)
	if err != nil {
		return err
	}

	client, err := utils.GetKongClient(config)
	if err != nil {
		return err
	}

	if targetContent.Info != nil {
		dumpConfig.SelectorTags = targetContent.Info.SelectorTags
	}

	// read the current state
	var currentState *state.KongState
	if workspaceExists {
		rawState, err := dump.Get(client, dumpConfig)
		if err != nil {
			return err
		}

		currentState, err = state.Get(rawState)
		if err != nil {
			return err
		}
	} else {
		print.CreatePrintln("creating workspace", targetContent.Workspace)

		// inject empty state
		currentState, err = state.NewKongState()
		if err != nil {
			return err
		}

		if !dry {
			_, err = rootClient.Workspaces.Create(nil, &kong.Workspace{Name: &targetContent.Workspace})
			if err != nil {
				return err
			}
		}
	}

	// read the target state
	rawState, err := file.Get(targetContent, file.RenderConfig{
		CurrentState: currentState,
		KongVersion:  kongVersion,
	})
	if err != nil {
		return err
	}
	targetState, err := state.Get(rawState)
	if err != nil {
		return err
	}

	s, _ := diff.NewSyncer(currentState, targetState)
	s.StageDelaySec = delay
	stats, errs := solver.Solve(stopChannel, s, client, parallelism, dry)
	if errs != nil {
		return utils.ErrArray{Errors: errs}
	}
	printFn := color.New(color.FgGreen, color.Bold).PrintfFunc()
	printFn("Summary:\n")
	printFn("  Created: %v\n", stats.CreateOps)
	printFn("  Updated: %v\n", stats.UpdateOps)
	printFn("  Deleted: %v\n", stats.DeleteOps)
	if diffCmdNonZeroExitCode &&
		stats.CreateOps+stats.UpdateOps+stats.DeleteOps != 0 {
		os.Exit(exitCodeDiffDetection)
	}
	return nil
}

func kongVersion(config utils.KongClientConfig) (semver.Version, error) {
	client, err := utils.GetKongClient(config)
	if err != nil {
		return semver.Version{}, err
	}

	root, err := client.Root(nil)
	if err != nil {
		return semver.Version{}, err
	}

	v, err := utils.CleanKongVersion(root["version"].(string))
	if err != nil {
		return semver.Version{}, err
	}
	return semver.ParseTolerant(v)
}

func validateNoArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return errors.New("positional arguments are not valid for this command, please use flags instead\n" +
			"Run 'deck --help' for usage.")
	}
	return nil
}
