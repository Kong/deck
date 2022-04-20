package cmd

import (
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
			cmdStr := "ping"
			var version string
			var err error
			mode := getMode(nil)
			if mode == modeKonnect {
				// get Konnect client
				httpClient := utils.HTTPClient()
				konnectClient, err := utils.GetKonnectClient(httpClient, konnectConfig)
				if err != nil {
					return err
				}
				u, _ := url.Parse(konnectConfig.Address)
				// authenticate with konnect
				res, err := authenticate(ctx, konnectClient, u.Host)
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
			} else {
				wsConfig := rootConfig.ForWorkspace(pingWorkspace)
				version, err = fetchKongVersion(ctx, wsConfig)
				if err != nil {
					return fmt.Errorf("reading Kong version: %w", err)
				}
				fmt.Println("Successfully connected to Kong!")
				fmt.Println("Kong version: ", version)
			}
			_ = sendAnalytics(cmdStr, version, mode)
			return nil
		},
	}

	pingCmd.Flags().StringVarP(&pingWorkspace, "workspace", "w",
		"", "Ping configuration with a specific Workspace "+
			"(Kong Enterprise only).\n"+
			"Useful when RBAC permissions are scoped to a Workspace.")
	return pingCmd
}
