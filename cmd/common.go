package cmd

import (
	"github.com/blang/semver"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/dump"
	"github.com/hbagdi/deck/file"
	"github.com/hbagdi/deck/solver"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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

// checkWorkspace checks if workspace exists in Kong.
func checkWorkspace(config utils.KongClientConfig) error {

	workspace := config.Workspace
	if workspace == "" {
		return nil
	}

	client, err := utils.GetKongClient(config)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET",
		utils.CleanAddress(config.Address)+"/workspaces/"+workspace,
		nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(nil, req, nil)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return errors.Errorf("workspace '%v' does not exist in Kong, "+
			"please create it before running decK.", workspace)
	}
	if err != nil {
		return errors.Wrapf(err, "failed to check workspace '%v' in Kong", workspace)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("unexpected status code while retrieving "+
			"workspace '%v' : %v", workspace, resp.StatusCode)
	}
	return nil
}

func syncMain(filenames []string, dry bool, parallelism, delay int) error {

	// read target file
	targetContent, err := file.GetContentFromFiles(filenames)
	if err != nil {
		return err
	}
	// prepare to read the current state from Kong
	config.Workspace = targetContent.Workspace

	// load Kong version before workspace
	kongVersion, err := kongVersion(config)
	if err != nil {
		return errors.Wrap(err, "reading Kong version")
	}

	if !config.SkipWorkspaceCrud {
		if err := checkWorkspace(config); err != nil {
			return err
		}
	}

	client, err := utils.GetKongClient(config)
	if err != nil {
		return err
	}

	if targetContent.Info != nil {
		dumpConfig.SelectorTags = targetContent.Info.SelectorTags
	}

	// read the current state
	rawState, err := dump.Get(client, dumpConfig)
	if err != nil {
		return err
	}
	currentState, err := state.Get(rawState)
	if err != nil {
		return err
	}

	// read the target state
	rawState, err = file.Get(targetContent, file.RenderConfig{
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

	var version string
	client, err := utils.GetKongClient(config)
	if err != nil {
		return semver.Version{}, err
	}
	root, err := client.Root(nil)
	if err != nil {
		// try with workspace path
		req, err := http.NewRequest("GET",
			utils.CleanAddress(config.Address)+"/"+config.Workspace+"/kong",
			nil)
		if err != nil {
			return semver.Version{}, err
		}
		var resp map[string]interface{}
		_, err = client.Do(nil, req, &resp)
		if err != nil {
			return semver.Version{}, err
		}
		version = resp["version"].(string)
	} else {
		version = root["version"].(string)
	}

	v, err := utils.CleanKongVersion(version)
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
