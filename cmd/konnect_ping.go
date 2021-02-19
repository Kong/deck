package cmd

import (
	"fmt"
	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// konnectPingCmd represents the ping2 command
var konnectPingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Verify connectivity with Konnect",
	Long: `Ping command can be used to verify if decK
can connect to Konnect's API endpoint. It also validates the supplied
credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := utils.GetKonnectClient(konnectConfig.Debug)
		if err != nil {
			return err
		}
		res, err := client.Auth.Login(cmd.Context(), konnectConfig.Email,
			konnectConfig.Password)
		if err != nil {
			return errors.Wrap(err, "authenticating with Konnect")
		}
		fmt.Printf("Successfully Konnected as %s %s (%s)!\n",
			res.FirstName, res.LastName, res.Organization)
		if konnectConfig.Debug {
			fmt.Printf("Organization ID: %s\n", res.OrganizationID)
		}
		return nil
	},
}

func init() {
	konnectCmd.AddCommand(konnectPingCmd)
}
