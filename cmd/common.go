package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/blang/semver/v4"
	"github.com/fatih/color"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/dump"
	"github.com/kong/deck/file"
	"github.com/kong/deck/print"
	"github.com/kong/deck/solver"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	exitCodeDiffDetection = 2
)

var (
	dumpConfig dump.Config
	assumeYes  bool
)

// workspaceExists checks if workspace exists in Kong.
func workspaceExists(ctx context.Context, rootConfig utils.KongClientConfig, workspaceName string) (bool, error) {
	if workspaceName == "" {
		// default workspace always exists
		return true, nil
	}

	if rootConfig.SkipWorkspaceCrud {
		// if RBAC user, skip check
		return true, nil
	}

	rootClient, err := utils.GetKongClient(rootConfig)
	if err != nil {
		return false, err
	}

	exists, err := rootClient.Workspaces.Exists(ctx, &workspaceName)
	if err != nil {
		return false, errors.Wrap(err, "checking if workspace exists")
	}
	return exists, nil
}

func syncMain(ctx context.Context, filenames []string, dry bool, parallelism,
	delay int, workspace string) error {

	// read target file
	targetContent, err := file.GetContentFromFiles(filenames)
	if err != nil {
		return err
	}

	rootClient, err := utils.GetKongClient(rootConfig)
	if err != nil {
		return err
	}

	var wsConfig utils.KongClientConfig
	var workspaceName string
	// prepare to read the current state from Kong
	if workspace != targetContent.Workspace && workspace != "" {
		print.DeletePrintf("Warning: Workspace '%v' specified via --workspace flag is "+
			"different from workspace '%v' found in state file(s).\n", workspace, targetContent.Workspace)
		workspaceName = workspace
	} else {
		workspaceName = targetContent.Workspace
	}
	wsConfig = rootConfig.ForWorkspace(workspaceName)

	// load Kong version after workspace
	kongVersion, err := fetchKongVersion(ctx, rootConfig)
	if err != nil {
		return errors.Wrap(err, "reading Kong version")
	}
	parsedKongVersion, err := parseKongVersion(kongVersion)
	if err != nil {
		return errors.Wrap(err, "parsing Kong version")
	}

	// TODO: instead of guessing the cobra command here, move the sendAnalytics
	// call to the RunE function. That is not trivial because it requires the
	// workspace name and kong client to be present on that level.
	cmd := "sync"
	if dry {
		cmd = "diff"
	}
	_ = sendAnalytics(cmd, kongVersion)

	workspaceExists, err := workspaceExists(ctx, rootConfig, workspaceName)
	if err != nil {
		return err
	}

	wsClient, err := utils.GetKongClient(wsConfig)
	if err != nil {
		return err
	}

	if targetContent.Info != nil {
		dumpConfig.SelectorTags = targetContent.Info.SelectorTags
	}

	// read the current state
	var currentState *state.KongState
	if workspaceExists {
		rawState, err := dump.Get(ctx, wsClient, dumpConfig)
		if err != nil {
			return err
		}

		currentState, err = state.Get(rawState)
		if err != nil {
			return err
		}
	} else {

		print.CreatePrintln("creating workspace", wsConfig.Workspace)

		// inject empty state
		currentState, err = state.NewKongState()
		if err != nil {
			return err
		}

		if !dry {
			_, err = rootClient.Workspaces.Create(nil, &kong.Workspace{Name: &wsConfig.Workspace})
			if err != nil {
				return err
			}
		}

	}

	// read the target state
	rawState, err := file.Get(targetContent, file.RenderConfig{
		CurrentState: currentState,
		KongVersion:  parsedKongVersion,
	})
	if err != nil {
		return err
	}
	if err := checkForRBACResources(*rawState, dumpConfig.RBACResourcesOnly); err != nil {
		return err
	}
	targetState, err := state.Get(rawState)
	if err != nil {
		return err
	}

	s, _ := diff.NewSyncer(currentState, targetState)
	s.StageDelaySec = delay
	stats, errs := solver.Solve(ctx, s, wsClient, nil, parallelism, dry)
	printFn := color.New(color.FgGreen, color.Bold).PrintfFunc()
	printFn("Summary:\n")
	printFn("  Created: %v\n", stats.CreateOps)
	printFn("  Updated: %v\n", stats.UpdateOps)
	printFn("  Deleted: %v\n", stats.DeleteOps)
	if errs != nil {
		return utils.ErrArray{Errors: errs}
	}
	if diffCmdNonZeroExitCode &&
		stats.CreateOps+stats.UpdateOps+stats.DeleteOps != 0 {
		os.Exit(exitCodeDiffDetection)
	}
	return nil
}

func fetchKongVersion(ctx context.Context, config utils.KongClientConfig) (string, error) {

	var version string

	workspace := config.Workspace

	// remove workspace to be able to call top-level / endpoint
	config.Workspace = ""
	client, err := utils.GetKongClient(config)
	if err != nil {
		return "", err
	}
	root, err := client.Root(ctx)
	if err != nil {
		if workspace == "" {
			return "", err
		}
		// try with workspace path
		req, err := http.NewRequest("GET",
			utils.CleanAddress(config.Address)+"/"+workspace+"/kong",
			nil)
		if err != nil {
			return "", err
		}
		var resp map[string]interface{}
		_, err = client.Do(ctx, req, &resp)
		if err != nil {
			return "", err
		}
		version = resp["version"].(string)
	} else {
		version = root["version"].(string)
	}
	return version, nil
}

func parseKongVersion(version string) (semver.Version, error) {
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

func checkForRBACResources(content utils.KongRawState,
	rbacResourcesOnly bool) error {
	proxyConfig := containsProxyConfiguration(content)
	rbacConfig := containsRBACConfiguration(content)
	if proxyConfig && rbacConfig {
		common := "At a time, state file(s) must entirely consist of either proxy " +
			"configuration or RBAC configuration."
		if rbacResourcesOnly {
			return fmt.Errorf("When --rbac-resources-only is used, state file(s) " +
				"cannot contain any resources other than RBAC resources. " + common)
		}
		return fmt.Errorf("State file(s) contains RBAC resources. " +
			"Please use --rbac-resources-only flag to manage these resources. " + common)
	}
	return nil
}

func containsProxyConfiguration(content utils.KongRawState) bool {
	return len(content.Services) != 0 ||
		len(content.Routes) != 0 ||
		len(content.Plugins) != 0 ||
		len(content.Upstreams) != 0 ||
		len(content.Certificates) != 0 ||
		len(content.CACertificates) != 0 ||
		len(content.Consumers) != 0
}

func containsRBACConfiguration(content utils.KongRawState) bool {
	return len(content.RBACRoles) != 0
}

func sendAnalytics(cmd, kongVersion string) error {
	if disableAnalytics {
		return nil
	}
	return utils.SendAnalytics(cmd, VERSION, kongVersion)
}
