package cmd

import (
	"fmt"
	"github.com/hbagdi/deck/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	pingWorkspace string
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Verify connectivity with Kong",
	Long: `Ping command can be used to verify if decK
can connect to Kong's Admin API or not.`,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := utils.GetKongClient(config)
		if err != nil {
			return errors.Wrap(err, "creating kong client")
		}
		conf, err := client.Root(nil)
		if err != nil {
			config.Workspace = pingWorkspace
			version, err := kongVersion(config)
			if err != nil {
				return errors.Wrap(err, "reading Kong version")
			}
			fmt.Println("Successfully connected to Kong!")
			fmt.Println("Kong version: ", version)
			return nil
		}
		version := conf["version"]
		if version == nil {
			return errors.New("version is nil from Kong")
		}
		fmt.Println("Successfully connected to Kong!")
		fmt.Println("Kong version: ", version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
	pingCmd.Flags().StringVarP(&pingWorkspace, "workspace", "w",
		"", "Ping configuration with a specific workspace "+
			"(Kong Enterprise only).\n"+
			"Useful when RBAC permissions are scoped to a workspace.")
}
