package cmd

import "github.com/hbagdi/deck/dump"

var stopChannel chan struct{}

// SetStopCh sets the stop channel for long running commands.
// This is useful for cases when a process needs to be cancelled gracefully
// before it can complete to finish. Example: SIGINT
func SetStopCh(stopCh chan struct{}) {
	stopChannel = stopCh
}

var dumpConfig dump.Config
