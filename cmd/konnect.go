package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var konnectAlphaState = `

WARNING: This command is currently in alpha state. This command
might have breaking changes in future releases.`

//nolint:errcheck
// newKonnectCmd represents the konnect command
func newKonnectCmd() *cobra.Command {
	konnectCmd := &cobra.Command{
		Use:   "konnect",
		Short: "Configuration tool for Konnect (in alpha)",
		Long: `The konnect command prints subcommands that can be used to
configure Konnect.` + konnectAlphaState,
	}
	// konnect-specific flags
	konnectCmd.PersistentFlags().String("konnect-email", "",
		"Email address associated with your Konnect account.")
	viper.BindPFlag("konnect-email",
		konnectCmd.PersistentFlags().Lookup("konnect-email"))

	konnectCmd.PersistentFlags().String("konnect-password", "",
		"Password associated with your Konnect account, "+
			"this takes precedence over --konnect-password-file flag.")
	viper.BindPFlag("konnect-password",
		konnectCmd.PersistentFlags().Lookup("konnect-password"))

	konnectCmd.PersistentFlags().String("konnect-password-file", "",
		"File containing the password to your Konnect account.")
	viper.BindPFlag("konnect-password-file",
		konnectCmd.PersistentFlags().Lookup("konnect-password-file"))

	konnectCmd.PersistentFlags().String("konnect-addr", "https://konnect.konghq.com",
		"Address of the Konnect endpoint.")
	viper.BindPFlag("konnect-addr",
		konnectCmd.PersistentFlags().Lookup("konnect-addr"))

	konnectCmd.AddCommand(newKonnectSyncCmd())
	konnectCmd.AddCommand(newKonnectPingCmd())
	konnectCmd.AddCommand(newKonnectDumpCmd())
	konnectCmd.AddCommand(newKonnectDiffCmd())
	return konnectCmd
}
