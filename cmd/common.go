package cmd

import (
	"context"
	"encoding/json"
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
	defaultFormatVersion  = "1.1"
	formatVersion30       = "3.0"
)

var (
	dumpConfig   dump.Config
	assumeYes    bool
	noMaskValues bool
)

type mode int

const (
	modeKonnect = iota
	modeKong
	modeKongEnterprise
)

type summary struct 
{
	Creating int32 `json:"creating"`
	Updating int32 `json:"updating"`
	Deleting int32 `json:"deleting"`
	Total 	 int32 `json:"total"`
}

type jsonOutput struct 
{
	Changes diff.EntityChanges `json:"changes"`
	Summary summary `json:"summary"`
	Warnings []string `json:"warnings"`
	Errors []string `json:"errors"`
}

var jsonOut jsonOutput
var isJsonOut bool

func getMode(targetContent *file.Content) mode {
	if inKonnectMode(targetContent) {
		return modeKonnect
	}
	return modeKong
}

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
		warning := fmt.Sprintf("Workspace '%v' specified via --workspace flag is "+
		"different from workspace '%v' found in state file(s).", workspaceFlag, targetContent.Workspace)
		if(isJsonOut){
			jsonOut.Warnings = append(jsonOut.Warnings, warning)
		}else{
			cprint.DeletePrintf("Warning: "+warning+"\n")
		}
		return workspaceFlag
	}
	return targetContent.Workspace
}

func syncMain(ctx context.Context, filenames []string, dry bool, parallelism,
	delay int, workspace string, isJsonOutput bool,
) error {
	// read target file
	if(isJsonOutput){
		isJsonOut = true
		jsonOut.Errors = []string{}
		jsonOut.Warnings = []string{}
		jsonOut.Changes = diff.EntityChanges{
			Creating: []diff.EntityState{},
			Updating: []diff.EntityState{},
			Deleting: []diff.EntityState{},
		}
	}
	targetContent, err := file.GetContentFromFiles(filenames)
	if err != nil {
		return err
	}
	if dumpConfig.SkipConsumers {
		targetContent.Consumers = []file.FConsumer{}
		targetContent.ConsumerGroups = []file.FConsumerGroupObject{}
	}
	if dumpConfig.SkipCACerts {
		targetContent.CACertificates = []file.FCACertificate{}
	}

	cmd := "sync"
	if dry {
		cmd = "diff"
	}

	var kongClient *kong.Client
	mode := getMode(targetContent)
	if mode == modeKonnect {
		if targetContent.Workspace != "" {
			return fmt.Errorf("_workspace set in config file.\n"+
				"Workspaces are not supported in Konnect. "+
				"Please remove '_workspace: %s' from your "+
				"configuration and try again", targetContent.Workspace)
		}
		if workspace != "" {
			return fmt.Errorf("--workspace flag is not supported when running against Konnect")
		}
		if targetContent.Konnect != nil {
			if konnectRuntimeGroup != "" &&
				targetContent.Konnect.RuntimeGroupName != konnectRuntimeGroup {
				return fmt.Errorf("warning: runtime group '%v' specified via "+
					"--konnect-runtime-group flag is "+
					"different from '%v' found in state file(s)",
					konnectRuntimeGroup, targetContent.Konnect.RuntimeGroupName)
			}
			konnectRuntimeGroup = targetContent.Konnect.RuntimeGroupName
		}
		kongClient, err = GetKongClientForKonnectMode(ctx, &konnectConfig)
		if err != nil {
			return err
		}
		dumpConfig.KonnectRuntimeGroup = konnectRuntimeGroup
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
	var kongVersion string
	var parsedKongVersion semver.Version
	if mode == modeKonnect {
		kongVersion, err = fetchKonnectKongVersion(ctx, kongClient)
		if err != nil {
			return fmt.Errorf("reading Konnect Kong version: %w", err)
		}
	} else {
		kongVersion, err = fetchKongVersion(ctx, wsConfig)
		if err != nil {
			return fmt.Errorf("reading Kong version: %w", err)
		}
	}
	parsedKongVersion, err = utils.ParseKongVersion(kongVersion)
	if err != nil {
		return fmt.Errorf("parsing Kong version: %w", err)
	}

	if parsedKongVersion.GTE(utils.Kong300Version) &&
		targetContent.FormatVersion != formatVersion30 {
		formatVersion := targetContent.FormatVersion
		if formatVersion == "" {
			formatVersion = defaultFormatVersion
		}
		return fmt.Errorf(
			"cannot apply '%s' config format version to Kong version 3.0 or above.\n"+
				utils.UpgradeMessage, formatVersion)
	}

	// TODO: instead of guessing the cobra command here, move the sendAnalytics
	// call to the RunE function. That is not trivial because it requires the
	// workspace name and kong client to be present on that level.
	_ = sendAnalytics(cmd, kongVersion, mode)

	workspaceExists, err := workspaceExists(ctx, rootConfig, workspaceName)
	if err != nil {
		return err
	}

	if kongClient == nil {
		kongClient, err = utils.GetKongClient(wsConfig)
		if err != nil {
			return err
		}
	}

	dumpConfig.SelectorTags, err = determineSelectorTag(*targetContent, dumpConfig)
	if err != nil {
		return err
	}

	// read the current state
	var currentState *state.KongState
	if workspaceExists {
		currentState, err = fetchCurrentState(ctx, kongClient, dumpConfig)
		if err != nil {
			return err
		}
	} else {
		// inject empty state
		currentState, err = state.NewKongState()
		if err != nil {
			return err
		}

		if(isJsonOutput){
			workspace := diff.EntityState{
				Name: wsConfig.Workspace,
				Type: "Workspace",
			}
			jsonOut.Changes.Creating = append(jsonOut.Changes.Creating, workspace)
		}else{
			cprint.CreatePrintln("Creating workspace", wsConfig.Workspace)
		}
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
	}, dumpConfig, kongClient)
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

	totalOps, err := performDiff(
		ctx, currentState, targetState, dry, parallelism, delay, kongClient, mode == modeKonnect)
	if err != nil {

		if(isJsonOut){
			if errs, ok := err.(utils.ErrArray)
			ok {
				jsonOut.Errors = append(jsonOut.Errors, errs.ErrorList()...)
			} else {
				jsonOut.Errors = append(jsonOut.Errors, fmt.Sprintf("%v", err))
			}
		}else{
			return err
		}
	}
	if diffCmdNonZeroExitCode && totalOps > 0 {
		os.Exit(exitCodeDiffDetection)
	}
	if(isJsonOutput){
		jsonOutStr, jsonErr := json.MarshalIndent(jsonOut, "", " ")
		if jsonErr != nil {
			return jsonErr
		}
		cprint.CreatePrintf(string(jsonOutStr)+"\n")
	}
	return nil
}

func determineSelectorTag(targetContent file.Content, config dump.Config) ([]string, error) {
	if targetContent.Info != nil {
		if len(targetContent.Info.SelectorTags) > 0 {
			utils.RemoveDuplicates(&targetContent.Info.SelectorTags)
			if len(config.SelectorTags) > 0 {
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
	dry bool, parallelism int, delay int, client *kong.Client, isKonnect bool,
) (int, error) {
	s, err := diff.NewSyncer(diff.SyncerOpts{
		CurrentState:  currentState,
		TargetState:   targetState,
		KongClient:    client,
		StageDelaySec: delay,
		NoMaskValues:  noMaskValues,
		IsKonnect:     isKonnect,
	})
	if err != nil {
		return 0, err
	}

	stats, errs, changes := s.Solve(ctx, parallelism, dry, isJsonOut)
	// print stats before error to report completed operations
	if(!isJsonOut){
		printStats(stats)
	}
	if errs != nil {
		return 0, utils.ErrArray{Errors: errs}
	}
	totalOps := stats.CreateOps.Count() + stats.UpdateOps.Count() + stats.DeleteOps.Count()

	jsonOut.Changes = diff.EntityChanges{
		Creating: append(jsonOut.Changes.Creating, changes.Creating...),
		Updating: append(jsonOut.Changes.Updating, changes.Updating...),
		Deleting: append(jsonOut.Changes.Deleting, changes.Deleting...),
	}
	jsonOut.Summary = summary{
		Creating: stats.CreateOps.Count(),
		Updating: stats.UpdateOps.Count(),
		Deleting: stats.DeleteOps.Count(),
		Total: totalOps,
	}
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

func validateNoArgs(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("positional arguments are not valid for this command, " +
			"please use flags instead")
	}
	return nil
}

func checkForRBACResources(content utils.KongRawState,
	rbacResourcesOnly bool,
) error {
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

func sendAnalytics(cmd, kongVersion string, mode mode) error {
	if disableAnalytics {
		return nil
	}
	var modeStr string
	switch mode {
	case modeKong:
		modeStr = "kong"
	case modeKonnect:
		modeStr = "konnect"
	case modeKongEnterprise:
		modeStr = "enterprise"
	}
	return utils.SendAnalytics(cmd, VERSION, kongVersion, modeStr)
}

func inKonnectMode(targetContent *file.Content) bool {
	if targetContent != nil && targetContent.Konnect != nil {
		return true
	} else if rootConfig.Address != defaultKongURL {
		return false
	} else if konnectConfig.Email != "" ||
		konnectConfig.Password != "" ||
		konnectConfig.Token != "" {
		return true
	}
	return false
}
