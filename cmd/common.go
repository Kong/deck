package cmd

import (
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/dump"
	"github.com/hbagdi/deck/file"
	"github.com/hbagdi/deck/solver"
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
	targetState, selectTags, workspace, err :=
		file.GetStateFromFile(filename)
	if err != nil {
		return err
	}
	config.Workspace = workspace
	client, err := utils.GetKongClient(config)
	if err != nil {
		return err
	}
	dumpConfig.SelectorTags = selectTags
	currentState, err := dump.GetState(client, dumpConfig)
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
