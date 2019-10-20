package cmd

import (
	"net/http"

	"github.com/hbagdi/deck/dump"
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

	req, err := http.NewRequest("GET", config.Address+"/workspace/"+workspace,
		nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(nil, req, nil)
	if resp.StatusCode == 404 {
		return errors.Errorf("workspace '%v' does not exist in Kong, "+
			"please create it before running decK.", workspace)
	}
	if err != nil {
		return errors.Wrapf(err, "checking workspace '%v' in Kong", workspace)
	}
	return nil
}
