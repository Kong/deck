package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/kong/go-database-reconciler/pkg/state"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
	"github.com/spf13/cobra"
)

var (
	resetCmdForce      bool
	resetWorkspace     string
	resetAllWorkspaces bool
	resetJSONOutput    bool
)

func executeReset(cmd *cobra.Command, _ []string) error {
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
	workspaces, err = fetchWorkspaces(ctx, false)
	if err != nil {
		return err
	}
	err = performReset(ctx, workspaces, false)
	if err != nil {
		return err
	}
	return nil
}

// newResetCmd represents the reset command
func newResetCmd(deprecated bool) *cobra.Command {
	short := "Reset deletes all entities in Kong"
	execute := executeReset
	if deprecated {
		short = "[deprecated] use 'deck gateway reset' instead"
		execute = func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stderr, "Info: 'deck reset' functionality has moved to 'deck gateway reset' and will be removed\n"+
				"in a future MAJOR version of deck. Migration to 'deck gateway reset' is recommended.\n")
			return executeReset(cmd, args)
		}
	}

	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: short,
		Long: `The reset command deletes all entities in Kong's database.string.

Use this command with extreme care as it's equivalent to running
"kong migrations reset" on your Kong instance.

By default, this command will ask for confirmation.`,
		Args: validateNoArgs,
		RunE: execute,
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

func getClientForReset(
	ctx context.Context,
	rootConfig utils.KongClientConfig,
	isKonnect bool,
	konnectConfig *utils.KonnectConfig,
) (*kong.Client, error) {
	if isKonnect {
		konnectClient, err := GetKongClientForKonnectMode(ctx, konnectConfig)
		if err != nil {
			return nil, err
		}
		return konnectClient, nil
	}
	kongClient, err := utils.GetKongClient(rootConfig)
	if err != nil {
		return nil, err
	}
	return kongClient, nil
}

func performReset(ctx context.Context, workspaces []string, isKonnect bool) error {
	for _, ws := range workspaces {
		konnectConfig.WorkspaceName = ws
		if ws == "default" {
			konnectConfig.WorkspaceName = ""
		}
		client, err := getClientForReset(ctx, rootConfig.ForWorkspace(ws), isKonnect, &konnectConfig)
		if err != nil {
			return fmt.Errorf("getting client for workspace '%s': %w", ws, err)
		}

		currentState, err := fetchCurrentState(ctx, client, dumpConfig)
		if err != nil {
			return fmt.Errorf("fetching state for workspace '%s': %w", ws, err)
		}
		targetState, err := state.NewKongState()
		if err != nil {
			return err
		}
		// Perform the diff/reset
		_, err = performDiff(ctx, currentState, targetState, false, 10, 0, client, isKonnect, resetJSONOutput, ApplyTypeFull)
		if err != nil {
			return fmt.Errorf("resetting workspace '%s': %w", ws, err)
		}
	}
	return nil
}

func fetchWorkspaces(ctx context.Context, isKonnect bool) ([]string, error) {
	baseClient, err := getClientForReset(ctx, rootConfig, isKonnect, &konnectConfig)
	if err != nil {
		return nil, fmt.Errorf("getting initial client: %w", err)
	}

	var workspaces []string
	if resetAllWorkspaces {
		workspaces, err = listWorkspaces(ctx, baseClient)
		if err != nil {
			return nil, fmt.Errorf("listing workspaces: %w", err)
		}
	} else if resetWorkspace != "" {
		exists, err := workspaceExists(ctx, rootConfig, resetWorkspace, isKonnect)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("workspace '%v' does not exist", resetWorkspace)
		}
		workspaces = append(workspaces, resetWorkspace)
	} else {
		// No workspace provided: reset global entities only
		workspaces = append(workspaces, "")
	}
	return workspaces, nil
}
