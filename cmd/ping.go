package cmd

import (
	"context"
	"fmt"

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
	// authenticate with Konnect
	_, _, err := getKongClientForKonnectMode(ctx)
	if err != nil {
		return err
	}

	fullName := konnectAuthResponse.FullName
	if fullName == "" {
		fullName = fmt.Sprintf("%s %s", konnectAuthResponse.FirstName, konnectAuthResponse.LastName)
	}
	fmt.Printf("Successfully Konnected as %s (%s)!\n",
		fullName, konnectAuthResponse.Organization)
	if konnectConfig.Debug {
		fmt.Printf("Organization ID: %s\n", konnectAuthResponse.OrganizationID)
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
