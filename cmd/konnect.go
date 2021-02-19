package cmd

import (
	"github.com/spf13/cobra"
)

// konnectCmd represents the konnect command
var konnectCmd = &cobra.Command{
	Use:   "konnect",
	Short: "Configuration tool for Konnect",
	Long: `Konnect command contains sub-commands that can be used to declarativley
configure Konnect`,
}

func init() {
	rootCmd.AddCommand(konnectCmd)
}
