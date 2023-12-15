package cmd

import (
	"github.com/kong/deck/cprint"
	"github.com/spf13/cobra"
)

var konnectAlphaState = `

WARNING: This command was in alpha state and has been deprecated.
This command will be removed in a future release.`

// newKonnectCmd represents the konnect command
func newKonnectCmd() *cobra.Command {
	konnectCmd := &cobra.Command{
		Use:   "konnect",
		Short: "[deprecated] Configuration tool for Konnect",
		Long: `The konnect command prints subcommands that can be used to
configure Konnect.` + konnectAlphaState,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cprint.UpdatePrintf("Notice: The 'deck konnect' command has been deprecated as of v1.12. \n" +
				"Please use deck <cmd> instead if you would like to declaratively manage your \n" +
				"Kong gateway config with Konnect.\n")
		},
	}
	konnectCmd.AddCommand(newKonnectSyncCmd())
	konnectCmd.AddCommand(newKonnectPingCmd())
	konnectCmd.AddCommand(newKonnectDumpCmd())
	konnectCmd.AddCommand(newKonnectDiffCmd())
	return konnectCmd
}
