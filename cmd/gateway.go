package cmd

import (
	"github.com/spf13/cobra"
)

func newGatewaySubCmd() *cobra.Command {
	gatewaySubCmd := &cobra.Command{
		Use:   "gateway [sub-command]...",
		Short: "Subcommand to host the decK network operations",
		Long:  `Subcommand to host the decK network operations.`,
	}

	return gatewaySubCmd
}
