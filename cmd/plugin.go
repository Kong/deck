package cmd

import (
	"github.com/spf13/cobra"
)

func newPluginCmd() *cobra.Command {
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Lint Kong plugins",
		Long:  `The plugin command set allows you to lint custom Kong plugin Lua code locally.`,
	}

	pluginCmd.AddCommand(newPluginLintCmd())

	return pluginCmd
}
