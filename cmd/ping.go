package cmd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/kong/deck/utils"
	"github.com/spf13/cobra"
)

var pingWorkspace string

// newPingCmd represents the ping command
func newPingCmd() *cobra.Command {
	pingCmd := &cobra.Command{
		Use:   "ping",
		Short: "Verify connectivity with Kong",
		Long: `The ping command can be used to verify if decK
can connect to Kong's Admin API.`,
		Args: validateNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			mode := getMode(nil)
			if mode == modeKonnect {
				return pingKonnect(ctx)
			}
			return pingKong(ctx)
		},
	}

	pingCmd.Flags().StringVarP(&pingWorkspace, "workspace", "w",
		"", "Ping configuration with a specific Workspace "+
			"(Kong Enterprise only).\n"+
			"Useful when RBAC permissions are scoped to a Workspace.")
	return pingCmd
}

func pingKonnect(ctx context.Context) error {
	// get Konnect client
	httpClient := utils.HTTPClient()
	kongClient, err := GetKongClientForKonnectMode(ctx, &konnectConfig)
	if err != nil {
		return err
	}
	konnectClient, err := utils.GetKonnectClient(httpClient, konnectConfig)
	if err != nil {
		return err
	}
	u, _ := url.Parse(konnectConfig.Address)
	if konnectConfig.Dev {
		version, err := fetchKonnectKongVersion(ctx, kongClient)
		if err != nil {
			return fmt.Errorf("reading Konnect Kong version: %w", err)
		}
		fmt.Println("Successfully Konnected to Kong!")
		fmt.Println("Kong version: ", version)
	} else {
		// authenticate with konnect
		res, err := authenticate(ctx, konnectClient, u.Host, konnectConfig)
		if err != nil {
			return fmt.Errorf("authenticating with Konnect: %w", err)
		}
		fullName := res.FullName
		if res.FullName == "" {
			fullName = fmt.Sprintf("%s %s", res.FirstName, res.LastName)
		}
		fmt.Printf("Successfully Konnected as %s (%s)!\n",
			fullName, res.Organization)
		if konnectConfig.Debug {
			fmt.Printf("Organization ID: %s\n", res.OrganizationID)
		}
	}
	_ = sendAnalytics("ping", "", modeKonnect)
	return nil
}

func pingKong(ctx context.Context) error {
	wsConfig := rootConfig.ForWorkspace(pingWorkspace)
	version, err := fetchKongVersion(ctx, wsConfig)
	if err != nil {
		return fmt.Errorf("reading Kong version: %w", err)
	}
	fmt.Println("Successfully connected to Kong!")
	fmt.Println("Kong version: ", version)
	_ = sendAnalytics("ping", version, modeKong)
	return nil
}
