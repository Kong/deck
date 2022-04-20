package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/kong/deck/diff"
	"github.com/kong/deck/dump"
	"github.com/kong/deck/file"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"golang.org/x/sync/errgroup"
)

const (
	defaultRuntimeGroupName        = "default"
	konnectWithRuntimeGroupsDomain = "api.konghq"
)

func authenticate(ctx context.Context, client *konnect.Client, host string) (konnect.AuthResponse, error) {
	if strings.Contains(host, konnectWithRuntimeGroupsDomain) {
		return client.Auth.LoginV2(ctx, konnectConfig.Email,
			konnectConfig.Password)
	}
	return client.Auth.Login(ctx, konnectConfig.Email,
		konnectConfig.Password)
}

func getKonnectClient(ctx context.Context) (*kong.Client, error) {
	httpClient := utils.HTTPClient()
	// get Konnect client
	konnectClient, err := utils.GetKonnectClient(httpClient, konnectConfig)
	if err != nil {
		return nil, err
	}

	var address string
	u, _ := url.Parse(konnectConfig.Address)
	// authenticate with konnect
	if _, err := authenticate(ctx, konnectClient, u.Host); err != nil {
		return nil, fmt.Errorf("authenticating with Konnect: %w", err)
	}
	if strings.Contains(u.Host, konnectWithRuntimeGroupsDomain) {
		// get kong runtime group ID
		kongRGID, err := fetchKongRuntimeGroupID(ctx, konnectClient)
		if err != nil {
			return nil, err
		}

		// set the kong runtime group ID in the client
		konnectClient.SetRuntimeGroupID(kongRGID)
		address = konnectConfig.Address + "/konnect-api/api/runtime_groups/" + kongRGID
	} else {
		// get kong control plane ID
		kongCPID, err := fetchKongControlPlaneID(ctx, konnectClient)
		if err != nil {
			return nil, err
		}

		// set the kong control plane ID in the client
		konnectClient.SetControlPlaneID(kongCPID)
		address = konnectConfig.Address + "/api/control_planes/" + kongCPID
	}
	// initialize kong client
	return utils.GetKongClient(utils.KongClientConfig{
		Address:    address,
		HTTPClient: httpClient,
		Debug:      konnectConfig.Debug,
	})
}

func resetKonnectV2(ctx context.Context) error {
	client, err := getKonnectClient(ctx)
	if err != nil {
		return err
	}
	currentState, err := fetchCurrentState(ctx, client, dumpConfig)
	if err != nil {
		return err
	}
	targetState, err := state.NewKongState()
	if err != nil {
		return err
	}
	_, err = performDiff(ctx, currentState, targetState, false, 10, 0, client)
	if err != nil {
		return err
	}
	return nil
}

func dumpKonnectV2(ctx context.Context) error {
	client, err := getKonnectClient(ctx)
	if err != nil {
		return err
	}
	if dumpCmdKongStateFile == "-" {
		return fmt.Errorf("writing to stdout is not supported in Konnect mode")
	}
	rawState, err := dump.Get(ctx, client, dumpConfig)
	if err != nil {
		return fmt.Errorf("reading configuration from Kong: %w", err)
	}
	ks, err := state.Get(rawState)
	if err != nil {
		return fmt.Errorf("building state: %w", err)
	}
	return file.KongStateToFile(ks, file.WriteConfig{
		Filename:         dumpCmdKongStateFile,
		FileFormat:       file.Format(strings.ToUpper(konnectDumpCmdStateFormat)),
		WithID:           dumpWithID,
		RuntimeGroupName: konnectRuntimeGroup,
	})
}

func syncKonnect(ctx context.Context,
	filenames []string, dry bool, parallelism int,
) error {
	httpClient := utils.HTTPClient()

	// read target file
	targetContent, err := file.GetContentFromFiles(filenames)
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
	})
	if err != nil {
		return err
	}

	stats, errs := s.Solve(ctx, parallelism, dry)
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

func fetchKongRuntimeGroupID(ctx context.Context,
	client *konnect.Client,
) (string, error) {
	runtimeGroups, _, err := client.RuntimeGroups.List(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("fetching runtime groups: %w", err)
	}
	if konnectRuntimeGroup == "" {
		konnectRuntimeGroup = defaultRuntimeGroupName
	}
	for _, rg := range runtimeGroups {
		if *rg.Name == konnectRuntimeGroup {
			return *rg.ID, nil
		}
	}
	return "", fmt.Errorf("runtime groups not found: %s", konnectRuntimeGroup)
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
