package cmd

import (
	"github.com/spf13/cobra"
)

var konnectAlphaState = `

WARNING: This command is currently in alpha state. This command
might have breaking changes in future releases.`

// konnectCmd represents the konnect command
var konnectCmd = &cobra.Command{
	Use:   "konnect",
	Short: "Configuration tool for Konnect (in alpha)",
	Long: `The konnect command prints sub-commands that can be used to
configure Konnect.` + konnectAlphaState,
}

func init() {
	rootCmd.AddCommand(konnectCmd)
}
