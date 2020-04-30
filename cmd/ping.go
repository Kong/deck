package cmd

import (
	"fmt"

	"github.com/hbagdi/deck/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Verify connectivity with Kong",
	Long: `Ping command can be used to verify if decK
can connect to Kong's Admin API or not.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("ping command cannot take any positional arguments. " +
				"Try using a flag instead.\n" +
				"For usage information: deck ping -h")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := utils.GetKongClient(config)
		if err != nil {
			return errors.Wrap(err, "creating kong client")
		}
		conf, err := client.Root(nil)
		if err != nil {
			return errors.Wrap(err, "connecting to kong")
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
}
