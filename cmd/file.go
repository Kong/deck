package cmd

import (
	"github.com/spf13/cobra"
)

func newAddFileCmd() *cobra.Command {
	addFileCmd := &cobra.Command{
		Use:   "file",
		Short: "Sub-command to host the decK file manipulation operations",
		Long:  `Sub-command to host the decK file manipulation operations`,
	}

	addFileCmd.AddCommand(newFileRenderCmd())

	return addFileCmd
}
