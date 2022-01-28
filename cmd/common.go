package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sort"

	"github.com/blang/semver/v4"
	"github.com/kong/deck/cprint"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/dump"
	"github.com/kong/deck/file"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
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
func workspaceExists(ctx context.Context, config utils.KongClientConfig, workspaceName string) (bool, error) {
	rootConfig := config.ForWorkspace("")
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

	exists, err := rootClient.Workspaces.ExistsByName(ctx, &workspaceName)
	if err != nil {
		return false, fmt.Errorf("checking if workspace exists: %w", err)
	}
	return exists, nil
}

func getWorkspaceName(workspaceFlag string, targetContent *file.Content) string {
	if workspaceFlag != targetContent.Workspace && workspaceFlag != "" {
		cprint.DeletePrintf("Warning: Workspace '%v' specified via --workspace flag is "+
			"different from workspace '%v' found in state file(s).\n", workspaceFlag, targetContent.Workspace)
		return workspaceFlag
	}
	return targetContent.Workspace
}

func syncMain(ctx context.Context, filenames []string, dry bool, parallelism,
	delay int, workspace string) error {

	// read target file
	targetContents, err := file.GetContentFromFiles(filenames)
	if err != nil {
		return err
	}

	for _, targetContent := range targetContents {
		if err := syncWs(ctx, targetContent, dry, parallelism, delay, workspace); err != nil {
			return err
		}
	}
	return nil
}

func syncWs(ctx context.Context, targetContent *file.Content, dry bool, parallelism,
	delay int, workspace string) error {
	if dumpConfig.SkipConsumers {
		targetContent.Consumers = []file.FConsumer{}
	}

	rootClient, err := utils.GetKongClient(rootConfig)
	if err != nil {
		return err
	}

	// prepare to read the current state from Kong
	var wsConfig utils.KongClientConfig
	workspaceName := getWorkspaceName(workspace, targetContent)
	wsConfig = rootConfig.ForWorkspace(workspaceName)

	// load Kong version after workspace
	kongVersion, err := fetchKongVersion(ctx, wsConfig)
	if err != nil {
		return fmt.Errorf("reading Kong version: %w", err)
	}
	parsedKongVersion, err := parseKongVersion(kongVersion)
	if err != nil {
		return fmt.Errorf("parsing Kong version: %w", err)
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

	dumpConfig.SelectorTags, err = determineSelectorTag(*targetContent, dumpConfig)
	if err != nil {
		return err
	}

	// read the current state
	var currentState *state.KongState
	if workspaceExists {
		currentState, err = fetchCurrentState(ctx, wsClient, dumpConfig)
		if err != nil {
			return err
		}
	} else {
		// inject empty state
		currentState, err = state.NewKongState()
		if err != nil {
			return err
		}

		cprint.CreatePrintln("creating workspace", wsConfig.Workspace)
		if !dry {
			_, err = rootClient.Workspaces.Create(ctx, &kong.Workspace{Name: &wsConfig.Workspace})
			if err != nil {
				return err
			}
		}
	}

	// read the target state
	rawState, err := file.Get(ctx, targetContent, file.RenderConfig{
		CurrentState: currentState,
		KongVersion:  parsedKongVersion,
	}, dumpConfig, wsClient)
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

	totalOps, err := performDiff(ctx, currentState, targetState, dry, parallelism, delay, wsClient, workspaceName)
	if err != nil {
		return err
	}

	if diffCmdNonZeroExitCode && totalOps > 0 {
		os.Exit(exitCodeDiffDetection)
	}
	return nil
}

func determineSelectorTag(targetContent file.Content, config dump.Config) ([]string, error) {
	if targetContent.Info != nil {
		if len(targetContent.Info.SelectorTags) > 0 {
			if len(config.SelectorTags) > 0 {
				utils.RemoveDuplicates(&targetContent.Info.SelectorTags)
				sort.Strings(config.SelectorTags)
				sort.Strings(targetContent.Info.SelectorTags)
				if !reflect.DeepEqual(config.SelectorTags, targetContent.Info.SelectorTags) {
					return nil, fmt.Errorf(`tags specified in the state file (%v) and via --select-tags flag (%v) are different.
					decK expects tags to be specified in either via flag or via state file.
					In case both are specified, they must match`, targetContent.Info.SelectorTags, config.SelectorTags)
				}
				// Both present and equal, return whichever
				return targetContent.Info.SelectorTags, nil
			}
			// Only targetContent.Info.SelectorTags present
			return targetContent.Info.SelectorTags, nil
		}
	}
	// Either targetContent.Info or targetContent.Info.SelectorTags is empty, return config tags
	return config.SelectorTags, nil
}

func fetchCurrentState(ctx context.Context, client *kong.Client, dumpConfig dump.Config) (*state.KongState, error) {
	rawState, err := dump.Get(ctx, client, dumpConfig)
	if err != nil {
		return nil, err
	}

	currentState, err := state.Get(rawState)
	if err != nil {
		return nil, err
	}
	return currentState, nil
}

func performDiff(ctx context.Context, currentState, targetState *state.KongState,
	dry bool, parallelism int, delay int, client *kong.Client, workspaceName string) (int, error) {
	s, err := diff.NewSyncer(diff.SyncerOpts{
		CurrentState:  currentState,
		TargetState:   targetState,
		KongClient:    client,
		StageDelaySec: delay,
	})
	if err != nil {
		return 0, err
	}

	stats, errs := s.Solve(ctx, parallelism, dry)
	// print stats before error to report completed operations
	printStats(stats, workspaceName)
	if errs != nil {
		return 0, utils.ErrArray{Errors: errs}
	}
	totalOps := stats.CreateOps.Count() + stats.UpdateOps.Count() + stats.DeleteOps.Count()
	return int(totalOps), nil
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
		return fmt.Errorf("positional arguments are not valid for this command, " +
			"please use flags instead")
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
