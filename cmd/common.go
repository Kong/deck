package cmd

import (
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/dump"
	"github.com/hbagdi/deck/file"
	"github.com/hbagdi/deck/solver"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
)

var stopChannel chan struct{}

// SetStopCh sets the stop channel for long running commands.
// This is useful for cases when a process needs to be cancelled gracefully
// before it can complete to finish. Example: SIGINT
func SetStopCh(stopCh chan struct{}) {
	stopChannel = stopCh
}

var dumpConfig dump.Config

func sync(filename string, dry bool) error {
	// read target file
	targetContent, err := file.GetContentFromFile(filename)
	if err != nil {
		return err
	}
	// prepare to read the current state from Kong
	config.Workspace = targetContent.Workspace
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

	targetState, _, _, err := file.GetStateFromContent(targetContent)
	if err != nil {
		return err
	}
	s, _ := diff.NewSyncer(currentState, targetState)
	errs := solver.Solve(stopChannel, s, client, dry)
	if errs != nil {
		return utils.ErrArray{Errors: errs}
	}
	return nil
}
