package cmd

import (
	"github.com/spf13/cobra"
)

func newAddFileCmd() *cobra.Command {
	addFileCmd := &cobra.Command{
		Use:   "file [sub-command]...",
		Short: "Subcommand to host the decK file manipulation operations",
		Long:  `Subcommand to host the decK file manipulation operations.`,
	}

	return addFileCmd
}
