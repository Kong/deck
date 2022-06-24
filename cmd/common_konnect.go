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
	defaultLegacyKonnectURL = "https://konnect.konghq.com"

	defaultRuntimeGroupName        = "default"
	konnectWithRuntimeGroupsDomain = "api.konghq"
)

var addresses = []string{
	defaultKonnectURL,
	defaultLegacyKonnectURL,
}

func authenticate(ctx context.Context, client *konnect.Client, host string) (konnect.AuthResponse, error) {
	if strings.Contains(host, konnectWithRuntimeGroupsDomain) {
		return client.Auth.LoginV2(ctx, konnectConfig.Email,
			konnectConfig.Password)
	}
	return client.Auth.Login(ctx, konnectConfig.Email,
		konnectConfig.Password)
}

// getKongClientForKonnectMode abstracts the different cloud environments users
// may be using, creating a Konnect client with the proper attributes set.
// This also includes a fallback mechanism using an address pool to establish
// a session with Konnect, making the different cloud environments completely
// transparent to users.
func getKongClientForKonnectMode(ctx context.Context) (*kong.Client, *konnect.Client, error) {
	httpClient := utils.HTTPClient()
	if konnectConfig.Address != defaultKonnectURL {
		addresses = []string{konnectConfig.Address}
	}
	// authenticate with konnect
	var err error
	var konnectClient *konnect.Client
	var parsedAddress *url.URL
	var konnectAddress string
	for _, address := range addresses {
		// get Konnect client
		konnectConfig.Address = address
		konnectClient, err = utils.GetKonnectClient(httpClient, konnectConfig)
		if err != nil {
			return nil, nil, err
		}
		parsedAddress, err = url.Parse(address)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing %s address: %v", address, err)
		}
		_, err = authenticate(ctx, konnectClient, parsedAddress.Host)
		if err == nil {
			break
		}
		if konnect.IsUnauthorizedErr(err) {
			continue
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("authenticating with Konnect: %w", err)
	}
	if strings.Contains(parsedAddress.Host, konnectWithRuntimeGroupsDomain) {
		// get kong runtime group ID
		kongRGID, err := fetchKongRuntimeGroupID(ctx, konnectClient)
		if err != nil {
			return nil, nil, err
		}

		// set the kong runtime group ID in the client
		konnectClient.SetRuntimeGroupID(kongRGID)
		konnectAddress = konnectConfig.Address + "/konnect-api/api/runtime_groups/" + kongRGID
	} else {
		// get kong control plane ID
		kongCPID, err := fetchKongControlPlaneID(ctx, konnectClient)
		if err != nil {
			return nil, nil, err
		}

		// set the kong control plane ID in the client
		konnectClient.SetControlPlaneID(kongCPID)
		konnectAddress = konnectConfig.Address + "/api/control_planes/" + kongCPID
	}
	// initialize kong client
	kongClient, err := utils.GetKongClient(utils.KongClientConfig{
		Address:    konnectAddress,
		HTTPClient: httpClient,
		Debug:      konnectConfig.Debug,
		Headers:    konnectConfig.Headers,
		Retryable:  true,
	})
	return kongClient, konnectClient, err
}

func resetKonnectV2(ctx context.Context) error {
	client, _, err := getKongClientForKonnectMode(ctx)
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
	client, _, err := getKongClientForKonnectMode(ctx)
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
	kongClient, konnectClient, err := getKongClientForKonnectMode(ctx)
	if err != nil {
		return err
	}

	var entityID string
	if strings.Contains(konnectConfig.Address, konnectWithRuntimeGroupsDomain) {
		// get kong runtime group ID
		entityID, err = fetchKongRuntimeGroupID(ctx, konnectClient)
		if err != nil {
			return err
		}

		// set the kong runtime group and control plane IDs in the client
		konnectClient.SetRuntimeGroupID(entityID)
		konnectClient.SetControlPlaneID(entityID)
	} else {
		// get kong control plane ID
		entityID, err = fetchKongControlPlaneID(ctx, konnectClient)
		if err != nil {
			return err
		}

		// set the kong control plane ID in the client
		konnectClient.SetControlPlaneID(entityID)
	}

	currentState, err := getKonnectState(ctx, kongClient, konnectClient, entityID,
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

	targetState, err := state.GetKonnectState(targetKongState, targetKonnectState, excludeServiceVersions)
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

	ks, err := state.GetKonnectState(kongState, konnectState, false)
	if err != nil {
		return nil, fmt.Errorf("building state: %w", err)
	}
	return ks, nil
}
