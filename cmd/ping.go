// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Verify connectivity with Kong",
	Long: `Ping command can be used to verify if deck
can connect to Kong's Admin API or not.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := GetKongClient(config)
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
