package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kong/go-database-reconciler/pkg/diff"
	"github.com/kong/go-database-reconciler/pkg/dump"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/konnect"
	"github.com/kong/go-database-reconciler/pkg/state"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
	"golang.org/x/sync/errgroup"
)

const (
	defaultControlPlaneName         = "default"
	maxRetriesForAuth               = 5
	apiRateLimitExceededErrorString = "API rate limit exceeded"
)

func authenticate(
	ctx context.Context, client *konnect.Client, konnectConfig utils.KonnectConfig,
) (konnect.AuthResponse, error) {
	attempts := 0
	backoff := 200 * time.Millisecond

	for {
		authResponse, err := client.Auth.LoginV2(ctx, konnectConfig.Email, konnectConfig.Password, konnectConfig.Token)
		if err == nil {
			return authResponse, nil
		}

		if !strings.Contains(err.Error(), apiRateLimitExceededErrorString) {
			return authResponse, err
		}

		attempts++
		if attempts > maxRetriesForAuth {
			return authResponse, fmt.Errorf("maximum retries (%d) exceeded for authentication", maxRetriesForAuth)
		}

		time.Sleep(backoff)
		backoff *= 2
	}
}

// GetKongClientForKonnectMode abstracts the different cloud environments users
// may be using, creating a Konnect client with the proper attributes set.
// This also includes a fallback mechanism using an address pool to establish
// a session with Konnect, making the different cloud environments completely
// transparent to users.
func GetKongClientForKonnectMode(
	ctx context.Context, konnectConfig *utils.KonnectConfig,
) (*kong.Client, error) {
	httpClient, err := utils.HTTPClientWithTLSConfig(konnectConfig.TLSConfig)
	if err != nil {
		return nil, err
	}

	if konnectConfig.Token != "" {
		found := false
		for _, h := range konnectConfig.Headers {
			if strings.HasPrefix(h, "Authorization:") {
				found = true
				break
			}
		}
		if !found {
			konnectConfig.Headers = append(
				konnectConfig.Headers, "Authorization:Bearer "+konnectConfig.Token,
			)
		}
	}

	if konnectConfig.Address == "" {
		konnectConfig.Address = defaultKonnectURL
	}

	// authenticate with konnect
	var konnectClient *konnect.Client
	var konnectAddress string
	// get Konnect client
	konnectClient, err = utils.GetKonnectClient(httpClient, *konnectConfig)
	if err != nil {
		return nil, err
	}
	_, err = authenticate(ctx, konnectClient, *konnectConfig)
	if err != nil {
		return nil, fmt.Errorf("authenticating with Konnect: %w", err)
	}
	cpID, err := fetchKonnectControlPlaneID(ctx, konnectClient, konnectConfig.ControlPlaneName)
	if err != nil {
		return nil, err
	}

	// set the kong control plane ID in the client
	konnectClient.SetRuntimeGroupID(cpID)
	konnectAddress = konnectConfig.Address + "/v2/control-planes/" + cpID + "/core-entities"

	// initialize kong client
	return utils.GetKongClient(utils.KongClientConfig{
		Address:    konnectAddress,
		HTTPClient: httpClient,
		Debug:      konnectConfig.Debug,
		Headers:    konnectConfig.Headers,
		Retryable:  true,
		TLSConfig:  konnectConfig.TLSConfig,
		Workspace:  konnectConfig.WorkspaceName,
	})
}

func resetKonnectV2(ctx context.Context) error {
	if konnectRuntimeGroup != "" {
		konnectControlPlane = konnectRuntimeGroup
	}
	if konnectControlPlane == "" {
		konnectControlPlane = defaultControlPlaneName
	}
	dumpConfig.KonnectControlPlane = konnectControlPlane
	konnectConfig.TLSConfig = rootConfig.TLSConfig
	baseKonnectClient, err := GetKongClientForKonnectMode(ctx, &konnectConfig)

	var workspaces []string
	if resetAllWorkspaces {
		if err != nil {
			return fmt.Errorf("getting initial konnect client: %w", err)
		}
		workspaces, err = listWorkspaces(ctx, baseKonnectClient)
		if err != nil {
			return fmt.Errorf("listing Konnect workspaces: %w", err)
		}
	} else if resetWorkspace != "" {
		workspaces = append(workspaces, resetWorkspace)
	} else {
		// No workspace provided: reset global entities only
		workspaces = append(workspaces, "")
	}
	for _, ws := range workspaces {
		// Update the config for this specific workspace iteration
		exists, err := workspaceExists(ctx, rootConfig, ws, true)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("workspace '%v' does not exist in Konnect", ws)
		}

		konnectConfig.WorkspaceName = ws

		client, err := GetKongClientForKonnectMode(ctx, &konnectConfig)
		if err != nil {
			return fmt.Errorf("getting client for workspace '%s': %w", ws, err)
		}

		currentState, err := fetchCurrentState(ctx, client, dumpConfig)
		if err != nil {
			return fmt.Errorf("fetching state for workspace '%s': %w", ws, err)
		}

		targetState, err := state.NewKongState()
		if err != nil {
			return err
		}

		// Perform the diff/reset
		_, err = performDiff(ctx, currentState, targetState, false, 10, 0, client, true, resetJSONOutput, ApplyTypeFull)
		if err != nil {
			return fmt.Errorf("resetting workspace '%s': %w", ws, err)
		}
	}

	return nil
}

func dumpKonnectV2(ctx context.Context) error {
	if konnectRuntimeGroup != "" {
		konnectControlPlane = konnectRuntimeGroup
	}
	if konnectControlPlane == "" {
		konnectControlPlane = defaultControlPlaneName
	}
	dumpConfig.KonnectControlPlane = konnectControlPlane
	konnectConfig.TLSConfig = rootConfig.TLSConfig

	format := file.Format(strings.ToUpper(dumpCmdStateFormat))
	writeConfig := file.WriteConfig{
		SelectTags:       dumpConfig.SelectorTags,
		FileFormat:       format,
		WithID:           dumpWithID,
		ControlPlaneName: konnectControlPlane,
		KongVersion:      fetchKonnectKongVersion(),
		SanitizeContent:  dumpConfig.SanitizeContent,
	}

	var workspacesList []string

	if dumpWorkspace != "" {
		workspacesList = append(workspacesList, dumpWorkspace)
	}

	if dumpAllWorkspaces {
		baseClient, err := GetKongClientForKonnectMode(ctx, &konnectConfig)
		if err != nil {
			return err
		}
		workspaces, err := listWorkspaces(ctx, baseClient)
		if err != nil {
			return err
		}
		workspacesList = append(workspacesList, workspaces...)
	}

	for _, workspace := range workspacesList {
		konnectConfig.WorkspaceName = workspace
		exists, err := workspaceExists(ctx, rootConfig, workspace, true)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("workspace '%v' does not exist in Konnect", workspace)
		}
		wsKonnectClient, err := GetKongClientForKonnectMode(ctx, &konnectConfig)
		if err != nil {
			return fmt.Errorf("getting client for workspace %s: %w", workspace, err)
		}
		writeConfig.Workspace = workspace
		writeConfig.Filename = workspace
		if err := performKonnectDump(ctx, wsKonnectClient, workspace, writeConfig); err != nil {
			return err
		}
	}

	// No workspace (default)
	// If dumpWorkspace is empty, GetKongClientForKonnectMode builds the CP-level URL
	client, err := GetKongClientForKonnectMode(ctx, &konnectConfig)
	if err != nil {
		return err
	}

	return performKonnectDump(ctx, client, dumpWorkspace, writeConfig)
}

func performKonnectDump(ctx context.Context, client *kong.Client, wsName string, cfg file.WriteConfig) error {
	rawState, err := dump.Get(ctx, client, dumpConfig)
	if err != nil {
		return fmt.Errorf("reading configuration: %w", err)
	}
	ks, err := state.Get(rawState)
	if err != nil {
		return fmt.Errorf("building state: %w", err)
	}

	cfg.Workspace = wsName
	if dumpAllWorkspaces {
		cfg.Filename = wsName
	} else {
		cfg.Filename = dumpCmdKongStateFile
	}

	if cfg.SanitizeContent {
		return sanitizeContent(ctx, client, ks, cfg, true)
	}
	return file.KongStateToFile(ks, cfg)
}

func syncKonnect(ctx context.Context,
	filenames []string, dry bool, parallelism int,
) error {
	httpClient := utils.HTTPClient()

	// read target file
	targetContent, err := file.GetContentFromFiles(filenames, false)
	if err != nil {
		return err
	}

	err = targetContent.PopulateDocumentContent(filenames)
	if err != nil {
		return fmt.Errorf("reading documents: %w", err)
	}

	targetContent.StripLocalDocumentPath()

	// get Konnect client
	konnectClient, err := utils.GetKonnectClient(httpClient, konnectConfig)
	if err != nil {
		return err
	}

	// authenticate with konnect
	_, err = konnectClient.Auth.Login(ctx,
		konnectConfig.Email,
		konnectConfig.Password)
	if err != nil {
		return fmt.Errorf("authenticating with Konnect: %w", err)
	}

	// get kong control plane ID
	kongCPID, err := fetchKongControlPlaneID(ctx, konnectClient)
	if err != nil {
		return err
	}

	// set the kong control plane ID in the client
	konnectClient.SetControlPlaneID(kongCPID)

	// initialize kong client
	kongClient, err := utils.GetKongClient(utils.KongClientConfig{
		Address:    konnectConfig.Address + "/api/control_planes/" + kongCPID,
		HTTPClient: httpClient,
		Debug:      konnectConfig.Debug,
		Headers:    konnectConfig.Headers,
	})
	if err != nil {
		return err
	}

	currentState, err := getKonnectState(ctx, kongClient, konnectClient, kongCPID,
		!konnectDumpIncludeConsumers)
	if err != nil {
		return err
	}

	targetKongState, targetKonnectState, err := file.GetForKonnect(ctx, targetContent, file.RenderConfig{
		CurrentState: currentState,
	}, kongClient)
	if err != nil {
		return err
	}

	targetState, err := state.GetKonnectState(targetKongState, targetKonnectState)
	if err != nil {
		return err
	}

	s, err := diff.NewSyncer(diff.SyncerOpts{
		CurrentState:  currentState,
		TargetState:   targetState,
		KongClient:    kongClient,
		KonnectClient: konnectClient,
		NoMaskValues:  noMaskValues,
	})
	if err != nil {
		return err
	}

	stats, errs, _ := s.Solve(ctx, parallelism, dry, false)

	// print stats before error to report completed operations
	printStats(stats)
	if errs != nil {
		return utils.ErrArray{Errors: errs}
	}
	if diffCmdNonZeroExitCode &&
		stats.CreateOps.Count()+stats.UpdateOps.Count()+stats.DeleteOps.Count() != 0 {
		os.Exit(exitCodeDiffDetection)
	}

	return nil
}

func fetchKongControlPlaneID(ctx context.Context,
	client *konnect.Client,
) (string, error) {
	controlPlanes, _, err := client.ControlPlanes.List(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("fetching control planes: %w", err)
	}

	return singleOutKongCP(controlPlanes)
}

func fetchKonnectControlPlaneID(
	ctx context.Context,
	client *konnect.Client,
	konnectControlPlaneName string,
) (string, error) {
	var runtimeGroups []*konnect.RuntimeGroup
	var listOpt *konnect.ListOpt
	for {
		currentRuntimeGroups, next, err := client.RuntimeGroups.List(ctx, listOpt)
		if err != nil {
			return "", fmt.Errorf("fetching control planes: %w", err)
		}
		runtimeGroups = append(runtimeGroups, currentRuntimeGroups...)
		if next == nil {
			break
		}
		listOpt = next
	}
	if konnectControlPlaneName != "" {
		konnectControlPlane = konnectControlPlaneName
	}
	if konnectControlPlane == "" {
		konnectControlPlane = defaultControlPlaneName
	}
	for _, rg := range runtimeGroups {
		if *rg.Name == konnectControlPlane {
			return *rg.ID, nil
		}
	}
	return "", fmt.Errorf("control plane not found: %s", konnectControlPlane)
}

func singleOutKongCP(controlPlanes []konnect.ControlPlane) (string, error) {
	kongCPCount := 0
	kongCPID := ""
	for _, controlPlane := range controlPlanes {
		if controlPlane.Type != nil &&
			!utils.Empty(controlPlane.ID) &&
			!utils.Empty(controlPlane.Type.Name) &&
			*controlPlane.Type.Name == "kong-ee" {
			kongCPCount++
			kongCPID = *controlPlane.ID
		}
	}
	if kongCPCount == 0 {
		return "", fmt.Errorf("found no Kong EE control-planes")
	}
	if kongCPCount > 1 {
		return "", fmt.Errorf("found multiple Kong EE control-planes, " +
			"decK expected a single control-plane")
	}
	return kongCPID, nil
}

func getKonnectState(ctx context.Context,
	kongClient *kong.Client,
	konnectClient *konnect.Client,
	kongCPID string,
	skipConsumers bool,
) (*state.KongState, error) {
	var kongState *utils.KongRawState
	var konnectState *utils.KonnectRawState

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		var err error
		// get export of Kong resources
		kongState, err = dump.Get(ctx, kongClient, dump.Config{
			SkipConsumers: skipConsumers,
		})
		if err != nil {
			return fmt.Errorf("reading configuration from Kong: %w", err)
		}
		return nil
	})

	group.Go(func() error {
		// get export of Konnect resources
		var err error
		konnectState, err = dump.GetFromKonnect(ctx, konnectClient,
			dump.KonnectConfig{ControlPlaneID: kongCPID})
		if err != nil {
			return err
		}
		return nil
	})

	err := group.Wait()
	if err != nil {
		return nil, err
	}

	ks, err := state.GetKonnectState(kongState, konnectState)
	if err != nil {
		return nil, fmt.Errorf("building state: %w", err)
	}
	return ks, nil
}

func fetchKonnectKongVersion() string {
	// Returning an hardcoded version for now. decK only needs the version
	// to determine the format_version expected in the state file. Since
	// Konnect is always on the latest version, we can safely return the
	// latest version here and avoid making an extra and unnecessary request.
	return "3.5.0.0"
}
