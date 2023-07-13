package cmd

import (
	"github.com/spf13/cobra"
)

func newFileSubCmd() *cobra.Command {
	fileCmd := &cobra.Command{
		Use:   "file [sub-command]...",
		Short: "Subcommand to host the decK file operations",
		Long:  `Subcommand to host the decK file operations.`,
	}

	return fileCmd
}
