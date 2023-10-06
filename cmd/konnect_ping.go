package cmd

import (
	"fmt"

	"github.com/kong/deck/utils"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = sendAnalytics("konnect-ping", "", modeKonnect)
			client, err := utils.GetKonnectClient(nil, konnectConfig)
			if err != nil {
				return err
			}
			res, err := client.Auth.Login(cmd.Context(), konnectConfig.Email,
				konnectConfig.Password)
			if err != nil {
				return fmt.Errorf("authenticating with Konnect: %w", err)
			}
			fmt.Printf("Successfully Konnected as %s %s (%s)!\n",
				res.FirstName, res.LastName, res.Organization)
			if konnectConfig.Debug {
				fmt.Printf("Organization ID: %s\n", res.OrganizationID)
			}
			return nil
		},
	}
}
