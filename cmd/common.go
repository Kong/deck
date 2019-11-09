package cmd

import (
	"net/http"

	"github.com/fatih/color"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/dump"
	"github.com/hbagdi/deck/file"
	"github.com/hbagdi/deck/solver"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/pkg/errors"
)

var stopChannel chan struct{}

// SetStopCh sets the stop channel for long running commands.
// This is useful for cases when a process needs to be cancelled gracefully
// before it can complete to finish. Example: SIGINT
func SetStopCh(stopCh chan struct{}) {
	stopChannel = stopCh
}

var dumpConfig dump.Config

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

	req, err := http.NewRequest("GET", config.Address+"/workspaces/"+workspace,
		nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(nil, req, nil)
	if err != nil {
		return errors.Wrapf(err, "checking workspace '%v' in Kong", workspace)
	}
	if resp.StatusCode == 404 {
		return errors.Errorf("workspace '%v' does not exist in Kong, "+
			"please create it before running decK.", workspace)
	}
	if resp.StatusCode != 200 {
		return errors.Errorf("unexpected error code while retrieving "+
			"workspace '%v' : %v", workspace, resp.StatusCode)
	}
	return nil
}

func syncMain(filename string, dry bool, parallelism int) error {
	// read target file
	targetContent, err := file.GetContentFromFile(filename)
	if err != nil {
		return err
	}
	// prepare to read the current state from Kong
	config.Workspace = targetContent.Workspace

	if err := checkWorkspace(config); err != nil {
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
	rawState, err := dump.Get(client, dumpConfig)
	if err != nil {
		return err
	}
	currentState, err := state.Get(rawState)
	if err != nil {
		return err
	}

	// read the target state
	rawState, err = file.Get(targetContent, currentState)
	if err != nil {
		return err
	}
	targetState, err := state.Get(rawState)
	if err != nil {
		return err
	}

	s, _ := diff.NewSyncer(currentState, targetState)
	stats, errs := solver.Solve(stopChannel, s, client, parallelism, dry)
	if errs != nil {
		return utils.ErrArray{Errors: errs}
	}
	printFn := color.New(color.FgGreen, color.Bold).PrintfFunc()
	printFn("Summary:\n")
	printFn("  Created: %v\n", stats.CreateOps)
	printFn("  Updated: %v\n", stats.UpdateOps)
	printFn("  Deleted: %v\n", stats.DeleteOps)
	return nil
}
