// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"github.com/hbagdi/deck/dump"
	"github.com/hbagdi/deck/reset"
	"github.com/hbagdi/go-kong/kong"
	"github.com/spf13/cobra"
)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO add a confirmation prompt
		// also add a -f/--force flag to get around the prompt
		client, err := kong.NewClient(nil, nil)
		if err != nil {
			return err
		}
		state, err := dump.Get(client)
		if err != nil {
			return err
		}
		err = reset.Reset(state, client)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
