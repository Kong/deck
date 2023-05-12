package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/kong/deck/dump"
	"github.com/kong/deck/file"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
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

func authenticate(
	ctx context.Context, client *konnect.Client, host string, konnectConfig utils.KonnectConfig,
) (konnect.AuthResponse, error) {
	if strings.Contains(host, konnectWithRuntimeGroupsDomain) {
		return client.Auth.LoginV2(ctx, konnectConfig.Email,
			konnectConfig.Password, konnectConfig.Token)
	}
	return client.Auth.Login(ctx, konnectConfig.Email,
		konnectConfig.Password)
}

// GetKongClientForKonnectMode abstracts the different cloud environments users
// may be using, creating a Konnect client with the proper attributes set.
// This also includes a fallback mechanism using an address pool to establish
// a session with Konnect, making the different cloud environments completely
// transparent to users.
func GetKongClientForKonnectMode(
	ctx context.Context, konnectConfig *utils.KonnectConfig,
) (*kong.Client, error) {
	httpClient := utils.HTTPClient()
	if konnectConfig.Address != defaultKonnectURL {
		addresses = []string{konnectConfig.Address}
	}

	if konnectConfig.Token != "" {
		konnectConfig.Headers = append(
			konnectConfig.Headers, "Authorization:Bearer "+konnectConfig.Token,
		)
	}

	// authenticate with konnect
	var err error
	var konnectClient *konnect.Client
	var parsedAddress *url.URL
	var konnectAddress string
	for _, address := range addresses {
		// get Konnect client
		konnectConfig.Address = address
		konnectClient, err = utils.GetKonnectClient(httpClient, *konnectConfig)
		if err != nil {
			return nil, err
		}
		parsedAddress, err = url.Parse(address)
		if err != nil {
			return nil, fmt.Errorf("parsing %s address: %v", address, err)
		}
		_, err = authenticate(ctx, konnectClient, parsedAddress.Host, *konnectConfig)
		if err == nil {
			break
		}
		// Personal Access Token authentication is not supported with the
		// legacy Konnect, so we don't need to fallback in case of 401s.
		if konnect.IsUnauthorizedErr(err) && konnectConfig.Token != "" {
			return nil, fmt.Errorf("authenticating with Konnect: %w", err)
		}
		if konnect.IsUnauthorizedErr(err) {
			continue
		}
	}
	if err != nil {
		return nil, fmt.Errorf("authenticating with Konnect: %w", err)
	}
	if strings.Contains(parsedAddress.Host, konnectWithRuntimeGroupsDomain) {
		// get kong runtime group ID
		kongRGID, err := fetchKongRuntimeGroupID(ctx, konnectClient)
		if err != nil {
			return nil, err
		}

		// set the kong runtime group ID in the client
		konnectClient.SetRuntimeGroupID(kongRGID)
		konnectAddress = konnectConfig.Address + "/konnect-api/api/runtime_groups/" + kongRGID
	} else {
		// get kong control plane ID
		kongCPID, err := fetchKongControlPlaneID(ctx, konnectClient)
		if err != nil {
			return nil, err
		}

		// set the kong control plane ID in the client
		konnectClient.SetControlPlaneID(kongCPID)
		konnectAddress = konnectConfig.Address + "/api/control_planes/" + kongCPID
	}

	// initialize kong client
	return utils.GetKongClient(utils.KongClientConfig{
		Address:    konnectAddress,
		HTTPClient: httpClient,
		Debug:      konnectConfig.Debug,
		Headers:    konnectConfig.Headers,
		Retryable:  true,
	})
}

func resetKonnectV2(ctx context.Context) error {
	client, err := GetKongClientForKonnectMode(ctx, &konnectConfig)
	if err != nil {
		return err
	}
	if dumpConfig.KonnectRuntimeGroup == "" {
		dumpConfig.KonnectRuntimeGroup = defaultRuntimeGroupName
	}
	currentState, err := fetchCurrentState(ctx, client, dumpConfig)
	if err != nil {
		return err
	}
	targetState, err := state.NewKongState()
	if err != nil {
		return err
	}
	_, err = performDiff(ctx, currentState, targetState, false, 10, 0, client, true)
	if err != nil {
		return err
	}
	return nil
}

func dumpKonnectV2(ctx context.Context) error {
	client, err := GetKongClientForKonnectMode(ctx, &konnectConfig)
	if err != nil {
		return err
	}
	if dumpConfig.KonnectRuntimeGroup == "" {
		dumpConfig.KonnectRuntimeGroup = defaultRuntimeGroupName
	}
	kongVersion, err := fetchKonnectKongVersion(ctx, client)
	if err != nil {
		return fmt.Errorf("reading Konnect Kong version: %w", err)
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
		SelectTags:       dumpConfig.SelectorTags,
		Filename:         dumpCmdKongStateFile,
		FileFormat:       file.Format(strings.ToUpper(dumpCmdStateFormat)),
		WithID:           dumpWithID,
		RuntimeGroupName: konnectRuntimeGroup,
		KongVersion:      kongVersion,
	})
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
	var runtimeGroups []*konnect.RuntimeGroup
	var listOpt *konnect.ListOpt
	for {
		currentRuntimeGroups, next, err := client.RuntimeGroups.List(ctx, listOpt)
		if err != nil {
			return "", fmt.Errorf("fetching runtime groups: %w", err)
		}
		runtimeGroups = append(runtimeGroups, currentRuntimeGroups...)
		if next == nil {
			break
		}
		listOpt = next
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

func fetchKonnectKongVersion(ctx context.Context, client *kong.Client) (string, error) {
	req, err := http.NewRequest("GET",
		utils.CleanAddress(client.BaseRootURL())+"/v1", nil)
	if err != nil {
		return "", err
	}

	var resp map[string]interface{}
	_, err = client.Do(ctx, req, &resp)
	if err != nil {
		return "", err
	}
	return resp["version"].(string), nil
}
