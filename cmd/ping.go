package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pingWorkspace string

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Verify connectivity with Kong",
	Long: `The ping command can be used to verify if decK
can connect to Kong's Admin API.`,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		wsConfig := rootConfig.ForWorkspace(pingWorkspace)
		version, err := fetchKongVersion(ctx, wsConfig)
		if err != nil {
			return fmt.Errorf("reading Kong version: %w", err)
		}
		_ = sendAnalytics("ping", version)
		fmt.Println("Successfully connected to Kong!")
		fmt.Println("Kong version: ", version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
	pingCmd.Flags().StringVarP(&pingWorkspace, "workspace", "w",
		"", "Ping configuration with a specific Workspace "+
			"(Kong Enterprise only).\n"+
			"Useful when RBAC permissions are scoped to a Workspace.")
}
