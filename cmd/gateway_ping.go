package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/spf13/cobra"
)

var pingWorkspace string

func executePing(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	mode := getMode(nil)
	if mode == modeKonnect {
		return pingKonnect(ctx)
	}
	return pingKong(ctx)
}

// newPingCmd represents the ping command
func newPingCmd(deprecated bool) *cobra.Command {
	short := "Verify connectivity with Kong"
	execute := executePing
	if deprecated {
		short = "[deprecated] use 'deck gateway ping' instead"
		execute = func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stderr, "Info: 'deck ping' functionality has moved to 'deck gateway ping' and will be removed\n"+
				"in a future MAJOR version of deck. Migration to 'deck gateway ping' is recommended.\n")
			return executePing(cmd, args)
		}
	}

	pingCmd := &cobra.Command{
		Use:   "ping",
		Short: short,
		Long: `The ping command can be used to verify if decK
can connect to Kong's Admin API.`,
		Args: validateNoArgs,
		RunE: execute,
	}

	pingCmd.Flags().StringVarP(&pingWorkspace, "workspace", "w",
		"", "Ping configuration with a specific Workspace "+
			"(Kong Enterprise only).\n"+
			"Useful when RBAC permissions are scoped to a Workspace.")
	return pingCmd
}

func pingKonnect(ctx context.Context) error {
	konnectConfig.TLSConfig = rootConfig.TLSConfig
	_, err := GetKongClientForKonnectMode(ctx, &konnectConfig)
	if err != nil {
		return err
	}
	// get Konnect client
	httpClient, err := utils.HTTPClientWithTLSConfig(rootConfig.TLSConfig)
	if err != nil {
		return err
	}
	konnectClient, err := utils.GetKonnectClient(httpClient, konnectConfig)
	if err != nil {
		return err
	}
	// authenticate with konnect
	res, err := authenticate(ctx, konnectClient, konnectConfig)
	if err != nil {
		return fmt.Errorf("authenticating with Konnect: %w", err)
	}

	fmt.Printf("Successfully Konnected to the %s organization!\n", res.Name)
	if konnectConfig.Debug {
		fmt.Printf("Organization ID: %s\n", res.OrganizationID)
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
