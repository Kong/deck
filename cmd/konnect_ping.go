package cmd

import (
	"fmt"

	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/spf13/cobra"
)

// newKonnectPingCmd represents the ping2 command
func newKonnectPingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "Verify connectivity with Konnect (in alpha)",
		Long: `The konnect ping command can be used to verify if decK
can connect to Konnect's API endpoint. It also validates the supplied
credentials.` + konnectAlphaState,
		Args: validateNoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_ = sendAnalytics("konnect-ping", "", modeKonnect)
			client, err := utils.GetKonnectClient(nil, konnectConfig)
			if err != nil {
				return err
			}
			res, err := authenticate(cmd.Context(), client, konnectConfig.Token)
			if err != nil {
				return fmt.Errorf("authenticating with Konnect: %w", err)
			}
			fmt.Printf("Successfully Konnected to the %s organization!\n", res.Name)
			if konnectConfig.Debug {
				fmt.Printf("Organization ID: %s\n", res.OrganizationID)
			}
			return nil
		},
	}
}
