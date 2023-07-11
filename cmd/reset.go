package cmd

import (
	"fmt"

	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/spf13/cobra"
)

var (
	resetCmdForce      bool
	resetWorkspace     string
	resetAllWorkspaces bool
	resetJSONOutput    bool
)

// newResetCmd represents the reset command
func newResetCmd() *cobra.Command {
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset deletes all entities in Kong",
		Long: `The reset command deletes all entities in Kong's database.string.

Use this command with extreme care as it's equivalent to running
"kong migrations reset" on your Kong instance.

By default, this command will ask for confirmation.`,
		Args: validateNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if resetAllWorkspaces && resetWorkspace != "" {
				return fmt.Errorf("workspace cannot be specified with --all-workspace flag")
			}

			if !resetCmdForce {
				ok, err := utils.Confirm("This will delete all configuration from Kong's database." +
					"\n> Are you sure? ")
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}

			mode := getMode(nil)
			if mode == modeKonnect {
				_ = sendAnalytics("reset", "", mode)
				return resetKonnectV2(ctx)
			}

			rootClient, err := utils.GetKongClient(rootConfig)
			if err != nil {
				return err
			}

			kongVersion, err := fetchKongVersion(ctx, rootConfig.ForWorkspace(resetWorkspace))
			if err != nil {
				return fmt.Errorf("reading Kong version: %w", err)
			}
			parsedKongVersion, err := utils.ParseKongVersion(kongVersion)
			if err != nil {
				return fmt.Errorf("parsing Kong version: %w", err)
			}
			_ = sendAnalytics("reset", kongVersion, mode)

			if utils.Kong340Version.LTE(parsedKongVersion) {
				dumpConfig.IsConsumerGroupScopedPluginSupported = true
			}

			var workspaces []string
			// Kong OSS or default workspace
			if !resetAllWorkspaces && resetWorkspace == "" {
				workspaces = append(workspaces, "")
			}

			// Kong Enterprise
			if resetAllWorkspaces {
				workspaces, err = listWorkspaces(ctx, rootClient)
				if err != nil {
					return err
				}
			}
			if resetWorkspace != "" {
				exists, err := workspaceExists(ctx, rootConfig, resetWorkspace)
				if err != nil {
					return err
				}
				if !exists {
					return fmt.Errorf("workspace '%v' does not exist in Kong", resetWorkspace)
				}

				workspaces = append(workspaces, resetWorkspace)
			}

			for _, workspace := range workspaces {
				wsClient, err := utils.GetKongClient(rootConfig.ForWorkspace(workspace))
				if err != nil {
					return err
				}
				currentState, err := fetchCurrentState(ctx, wsClient, dumpConfig)
				if err != nil {
					return err
				}
				targetState, err := state.NewKongState()
				if err != nil {
					return err
				}
				_, err = performDiff(ctx, currentState, targetState, false, 10, 0, wsClient, false, resetJSONOutput)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	resetCmd.Flags().BoolVarP(&resetCmdForce, "force", "f",
		false, "Skip interactive confirmation prompt before reset.")
	resetCmd.Flags().BoolVar(&dumpConfig.SkipConsumers, "skip-consumers",
		false, "do not reset consumers, consumer-groups or "+
			"any plugins associated with consumers.")
	resetCmd.Flags().StringVarP(&resetWorkspace, "workspace", "w",
		"", "reset configuration of a specific workspace"+
			"(Kong Enterprise only).")
	resetCmd.Flags().BoolVar(&resetAllWorkspaces, "all-workspaces",
		false, "reset configuration of all workspaces (Kong Enterprise only).")
	resetCmd.Flags().BoolVar(&noMaskValues, "no-mask-deck-env-vars-value",
		false, "do not mask DECK_ environment variable values at diff output.")
	resetCmd.Flags().StringSliceVar(&dumpConfig.SelectorTags,
		"select-tag", []string{},
		"only entities matching tags specified via this flag are deleted.\n"+
			"When this setting has multiple tag values, entities must match every tag.")
	resetCmd.Flags().BoolVar(&dumpConfig.RBACResourcesOnly, "rbac-resources-only",
		false, "reset only the RBAC resources (Kong Enterprise only).")
	resetCmd.Flags().BoolVar(&dumpConfig.SkipCACerts, "skip-ca-certificates",
		false, "do not reset CA certificates.")
	resetCmd.Flags().BoolVar(&resetJSONOutput, "json-output",
		false, "generate command execution report in a JSON format")

	return resetCmd
}
