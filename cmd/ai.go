package cmd

import (
	"github.com/spf13/cobra"
)

func newAiSubCmd() *cobra.Command {
	aiSubCmd := &cobra.Command{
		Use:   "ai [sub-command]...",
		Short: "Subcommand to host decK AI Gateway operations",
		Long:  `Subcommand to host decK AI Gateway operations.`,
	}

	return aiSubCmd
}
