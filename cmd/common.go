package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sort"

	"github.com/blang/semver/v4"
	"github.com/kong/deck/utils"
	"github.com/kong/go-database-reconciler/pkg/cprint"
	"github.com/kong/go-database-reconciler/pkg/diff"
	"github.com/kong/go-database-reconciler/pkg/dump"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/state"
	reconcilerUtils "github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
	"github.com/spf13/cobra"
)

const exitCodeDiffDetection = 2

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
	modeLocal
)

var jsonOutput diff.JSONOutputObject

func getMode(targetContent *file.Content) mode {
	if inKonnectMode(targetContent) {
		return modeKonnect
	}
	return modeKong
}

// workspaceExists checks if workspace exists in Kong.
func workspaceExists(ctx context.Context, config reconcilerUtils.KongClientConfig, workspaceName string) (bool, error) {
	rootConfig := config.ForWorkspace("")
	if workspaceName == "" {
		// default workspace always exists
		return true, nil
	}

	if rootConfig.SkipWorkspaceCrud {
		// if RBAC user, skip check
		return true, nil
	}

	rootClient, err := reconcilerUtils.GetKongClient(rootConfig)
	if err != nil {
		return false, err
	}

	exists, err := rootClient.Workspaces.ExistsByName(ctx, &workspaceName)
	if err != nil {
		return false, fmt.Errorf("checking if workspace exists: %w", err)
	}
	return exists, nil
}

func getWorkspaceName(workspaceFlag string, targetContent *file.Content,
	enableJSONOutput bool,
) string {
	if workspaceFlag != targetContent.Workspace && workspaceFlag != "" {
		warning := fmt.Sprintf("Workspace '%v' specified via --workspace flag is "+
			"different from workspace '%v' found in state file(s).", workspaceFlag, targetContent.Workspace)
		if enableJSONOutput {
			jsonOutput.Warnings = append(jsonOutput.Warnings, warning)
		} else {
			cprint.DeletePrintf("Warning: " + warning + "\n")
		}
		return workspaceFlag
	}
	return targetContent.Workspace
}

func evaluateTargetRuntimeGroupOrControlPlaneName(targetContent *file.Content) error {
	targetControlPlane := targetContent.Konnect.ControlPlaneName
	targetRuntimeGroup := targetContent.Konnect.RuntimeGroupName
	if targetControlPlane != "" && targetRuntimeGroup != "" {
		return errors.New(`cannot set both runtime_group_name and control_plane_name. ` +
			`Please use only control_plane_name`)
	}
	targetFromFile := targetControlPlane
	if targetFromFile == "" {
		targetFromFile = targetRuntimeGroup
	}
	targetFromCLI := konnectControlPlane
	if targetFromCLI == "" {
		targetFromCLI = konnectRuntimeGroup
	}
	if targetFromCLI != "" && targetFromFile != targetFromCLI {
		return fmt.Errorf("warning: control plane '%v' specified via "+
			"--konnect-[control-plane|runtime-group]-name flag is "+
			"different from '%v' found in state file(s)",
			targetFromCLI, targetFromFile)
	}
	if targetControlPlane != "" {
		konnectControlPlane = targetControlPlane
	}
	if targetRuntimeGroup != "" {
		konnectControlPlane = targetRuntimeGroup
	}
	return nil
}

func syncMain(ctx context.Context, filenames []string, dry bool, parallelism,
	delay int, workspace string, enableJSONOutput bool,
) error {
	// read target file
	if enableJSONOutput {
		jsonOutput.Errors = []string{}
		jsonOutput.Warnings = []string{}
		jsonOutput.Changes = diff.EntityChanges{
			Creating: []diff.EntityState{},
			Updating: []diff.EntityState{},
			Deleting: []diff.EntityState{},
		}
	}
	targetContent, err := file.GetContentFromFiles(filenames, false)
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
			if err := evaluateTargetRuntimeGroupOrControlPlaneName(targetContent); err != nil {
				return err
			}
		}
		if konnectRuntimeGroup != "" {
			konnectControlPlane = konnectRuntimeGroup
		}
		konnectConfig.TLSConfig = rootConfig.TLSConfig
		kongClient, err = GetKongClientForKonnectMode(ctx, &konnectConfig)
		if err != nil {
			return err
		}
		dumpConfig.KonnectControlPlane = konnectControlPlane
	}

	rootClient, err := reconcilerUtils.GetKongClient(rootConfig)
	if err != nil {
		return err
	}

	// prepare to read the current state from Kong
	var wsConfig reconcilerUtils.KongClientConfig
	workspaceName := getWorkspaceName(workspace, targetContent, enableJSONOutput)
	wsConfig = rootConfig.ForWorkspace(workspaceName)

	// load Kong version after workspace
	var kongVersion string
	var parsedKongVersion semver.Version
	if mode == modeKonnect {
		kongVersion = fetchKonnectKongVersion()
	} else {
		kongVersion, err = fetchKongVersion(ctx, wsConfig)
		if err != nil {
			return fmt.Errorf("reading Kong version: %w", err)
		}
	}
	parsedKongVersion, err = reconcilerUtils.ParseKongVersion(kongVersion)
	if err != nil {
		return fmt.Errorf("parsing Kong version: %w", err)
	}

	if parsedKongVersion.GTE(reconcilerUtils.Kong300Version) &&
		targetContent.FormatVersion != utils.FormatVersion30 {
		formatVersion := targetContent.FormatVersion
		if formatVersion == "" {
			formatVersion = utils.DefaultFormatVersion
		}
		return fmt.Errorf(
			"cannot apply '%s' config format version to Kong version 3.0 or above.\n"+
				reconcilerUtils.UpgradeMessage, formatVersion)
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
		kongClient, err = reconcilerUtils.GetKongClient(wsConfig)
		if err != nil {
			return err
		}
	}

	dumpConfig.IsFilterChainsSupported, err = determineFilterChainSupport(ctx, kongClient)
	if err != nil {
		return err
	}

	dumpConfig.SelectorTags, err = determineSelectorTag(*targetContent, dumpConfig)
	if err != nil {
		return err
	}

	dumpConfig.LookUpSelectorTagsConsumers, err = determineLookUpSelectorTagsConsumers(*targetContent)
	if err != nil {
		return fmt.Errorf("error determining lookup selector tags for consumers: %w", err)
	}

	if dumpConfig.LookUpSelectorTagsConsumers != nil {
		consumersGlobal, err := dump.GetAllConsumers(ctx, kongClient, dumpConfig.LookUpSelectorTagsConsumers)
		if err != nil {
			return fmt.Errorf("error retrieving global consumers via lookup selector tags: %w", err)
		}
		for _, c := range consumersGlobal {
			targetContent.Consumers = append(targetContent.Consumers, file.FConsumer{Consumer: *c})
			if err != nil {
				return fmt.Errorf("error adding global consumer %v: %w", *c.Username, err)
			}
		}
	}

	dumpConfig.LookUpSelectorTagsRoutes, err = determineLookUpSelectorTagsRoutes(*targetContent)
	if err != nil {
		return fmt.Errorf("error determining lookup selector tags for routes: %w", err)
	}

	if dumpConfig.LookUpSelectorTagsRoutes != nil {
		routesGlobal, err := dump.GetAllRoutes(ctx, kongClient, dumpConfig.LookUpSelectorTagsRoutes)
		if err != nil {
			return fmt.Errorf("error retrieving global routes via lookup selector tags: %w", err)
		}
		for _, r := range routesGlobal {
			targetContent.Routes = append(targetContent.Routes, file.FRoute{Route: *r})
			if err != nil {
				return fmt.Errorf("error adding global route %v: %w", r.FriendlyName(), err)
			}
		}
	}

	if reconcilerUtils.Kong340Version.LTE(parsedKongVersion) {
		dumpConfig.IsConsumerGroupScopedPluginSupported = true
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

		if enableJSONOutput {
			workspace := diff.EntityState{
				Name: wsConfig.Workspace,
				Kind: "workspace",
			}
			jsonOutput.Changes.Creating = append(jsonOutput.Changes.Creating, workspace)
		} else {
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
	if err := checkFilterChainsAllowed(*rawState, dumpConfig, parsedKongVersion); err != nil {
		return err
	}
	targetState, err := state.Get(rawState)
	if err != nil {
		return err
	}

	totalOps, err := performDiff(
		ctx, currentState, targetState, dry, parallelism, delay, kongClient, mode == modeKonnect, enableJSONOutput)
	if err != nil {
		if enableJSONOutput {
			var errs reconcilerUtils.ErrArray
			if errors.As(err, &errs) {
				jsonOutput.Errors = append(jsonOutput.Errors, errs.ErrorList()...)
			} else {
				jsonOutput.Errors = append(jsonOutput.Errors, err.Error())
			}
		} else {
			return err
		}
	}
	if diffCmdNonZeroExitCode && totalOps > 0 {
		os.Exit(exitCodeDiffDetection)
	}
	if enableJSONOutput {
		jsonOutputBytes, jsonErr := json.MarshalIndent(jsonOutput, "", "\t")
		if jsonErr != nil {
			return err
		}
		jsonOutputString := string(jsonOutputBytes)
		if !noMaskValues {
			jsonOutputString = diff.MaskEnvVarValue(jsonOutputString)
		}

		cprint.BluePrintLn(jsonOutputString + "\n")
	}
	return nil
}

func determineLookUpSelectorTagsConsumers(targetContent file.Content) ([]string, error) {
	if targetContent.Info != nil &&
		targetContent.Info.LookUpSelectorTags != nil &&
		targetContent.Info.LookUpSelectorTags.Consumers != nil {
		if len(targetContent.Info.LookUpSelectorTags.Consumers) == 0 {
			return nil, fmt.Errorf("global consumers specified but no global tags")
		}
		reconcilerUtils.RemoveDuplicates(&targetContent.Info.LookUpSelectorTags.Consumers)
		sort.Strings(targetContent.Info.LookUpSelectorTags.Consumers)
		return targetContent.Info.LookUpSelectorTags.Consumers, nil

	}
	return nil, nil
}

func determineLookUpSelectorTagsRoutes(targetContent file.Content) ([]string, error) {
	if targetContent.Info != nil &&
		targetContent.Info.LookUpSelectorTags != nil &&
		targetContent.Info.LookUpSelectorTags.Routes != nil {
		if len(targetContent.Info.LookUpSelectorTags.Routes) == 0 {
			return nil, fmt.Errorf("global routes specified but no global tags")
		}
		reconcilerUtils.RemoveDuplicates(&targetContent.Info.LookUpSelectorTags.Routes)
		sort.Strings(targetContent.Info.LookUpSelectorTags.Routes)
		return targetContent.Info.LookUpSelectorTags.Routes, nil

	}
	return nil, nil
}

func determineSelectorTag(targetContent file.Content, config dump.Config) ([]string, error) {
	if targetContent.Info != nil {
		if len(targetContent.Info.SelectorTags) > 0 {
			reconcilerUtils.RemoveDuplicates(&targetContent.Info.SelectorTags)
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

func determineFilterChainSupport(ctx context.Context, client *kong.Client) (bool, error) {
	workspace := client.Workspace()
	client.SetWorkspace("")
	defer client.SetWorkspace(workspace)

	data, err := client.Root(ctx)
	if err != nil {
		return false, fmt.Errorf("fetching gateway configuration: %w", err)
	}

	conf, ok := data["configuration"]
	if !ok {
		return false, errors.New("failed to extract gateway configuration")
	}

	confmap, ok := conf.(map[string]interface{})
	if !ok {
		return false, errors.New("failed to extract gateway configuration")
	}

	wasm, ok := confmap["wasm"]
	// probably just an old version of Kong
	if !ok {
		return false, nil
	}

	var supported bool
	supported, ok = wasm.(bool)
	if !ok {
		return false, errors.New("could not determine if wasm is enabled on this gateway node")
	}

	return supported, nil
}

func checkFilterChainsAllowed(
	content reconcilerUtils.KongRawState,
	dumpConfig dump.Config,
	version semver.Version,
) error {
	if len(content.FilterChains) < 1 {
		return nil
	} else if reconcilerUtils.Kong340Version.GT(version) {
		return errors.New(
			"one or more Wasm Filter Chain entities were found, but the target " +
				"Kong Gateway node does not support Wasm.\n" +
				"Wasm Filter Chain support requires Kong >= 3.4",
		)
	} else if !dumpConfig.IsFilterChainsSupported {
		return errors.New("one or more Wasm Filter Chain entities were found, " +
			"but Wasm is not enabled on the target Kong Gateway node",
		)
	}

	return nil
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
	enableJSONOutput bool,
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

	stats, errs, changes := s.Solve(ctx, parallelism, dry, enableJSONOutput)
	// print stats before error to report completed operations
	if !enableJSONOutput {
		printStats(stats)
	}
	if errs != nil {
		return 0, reconcilerUtils.ErrArray{Errors: errs}
	}
	totalOps := stats.CreateOps.Count() + stats.UpdateOps.Count() + stats.DeleteOps.Count()

	if enableJSONOutput {
		jsonOutput.Changes = diff.EntityChanges{
			Creating: append(jsonOutput.Changes.Creating, changes.Creating...),
			Updating: append(jsonOutput.Changes.Updating, changes.Updating...),
			Deleting: append(jsonOutput.Changes.Deleting, changes.Deleting...),
		}
		jsonOutput.Summary = diff.Summary{
			Creating: stats.CreateOps.Count(),
			Updating: stats.UpdateOps.Count(),
			Deleting: stats.DeleteOps.Count(),
			Total:    totalOps,
		}
	}
	return int(totalOps), nil
}

func fetchKongVersion(ctx context.Context, config reconcilerUtils.KongClientConfig) (string, error) {
	var version string

	workspace := config.Workspace

	// remove workspace to be able to call top-level / endpoint
	config.Workspace = ""
	client, err := reconcilerUtils.GetKongClient(config)
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
			reconcilerUtils.CleanAddress(config.Address)+"/"+workspace+"/kong",
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

func checkForRBACResources(content reconcilerUtils.KongRawState,
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

func containsProxyConfiguration(content reconcilerUtils.KongRawState) bool {
	return len(content.Services) != 0 ||
		len(content.Routes) != 0 ||
		len(content.Plugins) != 0 ||
		len(content.FilterChains) != 0 ||
		len(content.Upstreams) != 0 ||
		len(content.Certificates) != 0 ||
		len(content.CACertificates) != 0 ||
		len(content.Consumers) != 0
}

func containsRBACConfiguration(content reconcilerUtils.KongRawState) bool {
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
	case modeLocal:
		modeStr = "local"
	}
	return reconcilerUtils.SendAnalytics(cmd, VERSION, kongVersion, modeStr)
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
