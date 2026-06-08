package cmd

import (
	"github.com/spf13/cobra"
)

func newPluginCmd() *cobra.Command {
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage and validate Kong plugins",
		Long:  `The plugin command set allows you to manage, validate, and lint Kong plugins locally.`,
	}

	pluginCmd.AddCommand(newPluginLintCmd())

	return pluginCmd
}
