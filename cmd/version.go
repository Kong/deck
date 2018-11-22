// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var VERSION = "dev"
var COMMIT = "unknown"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of decK",
	Long: `version prints the version of deck along with git short
commit hash of the source tree`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("decK %s (%s) \n", VERSION, COMMIT)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
