package cmd

import (
	"github.com/kong/deck/dump"
	"github.com/kong/deck/reset"
	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	resetCmdForce      bool
	resetWorkspace     string
	resetAllWorkspaces bool
)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset deletes all entities in Kong",
	Long: `Reset command will delete all entities in Kong's database.string

Use this command with extreme care as it is equivalent to running
"kong migrations reset" on your Kong instance.

By default, this command will ask for a confirmation prompt.`,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		rootClient, err := utils.GetKongClient(rootConfig)
		if err != nil {
			return err
		}
		// Kong OSS or default workspace
		if !resetAllWorkspaces && resetWorkspace == "" {
			state, err := dump.Get(rootClient, dumpConfig)
			if err != nil {
				return err
			}
			err = reset.Reset(state, rootClient)
			if err != nil {
				return err
			}
			return nil
		}

		if resetAllWorkspaces && resetWorkspace != "" {
			return errors.New("workspace cannot be specified with --all-workspace flag")
		}

		// Kong Enterprise
		var workspaces []string
		if resetAllWorkspaces {
			workspaces, err = listWorkspaces(cmd.Context(), rootClient)
			if err != nil {
				return err
			}
		}
		if resetWorkspace != "" {
			exists, err := workspaceExists(rootConfig.ForWorkspace(resetWorkspace))
			if err != nil {
				return err
			}
			if !exists {
				return errors.Errorf("workspace '%v' does not exist in Kong", resetWorkspace)
			}

			workspaces = append(workspaces, resetWorkspace)
		}

		for _, workspace := range workspaces {
			wsClient, err := utils.GetKongClient(rootConfig.ForWorkspace(workspace))
			if err != nil {
				return err
			}
			state, err := dump.Get(wsClient, dumpConfig)
			if err != nil {
				return err
			}
			err = reset.Reset(state, wsClient)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
	resetCmd.Flags().BoolVarP(&resetCmdForce, "force", "f",
		false, "Skip interactive confirmation prompt before reset")
	resetCmd.Flags().BoolVar(&dumpConfig.SkipConsumers, "skip-consumers",
		false, "do not reset consumers or "+
			"any plugins associated with consumers")
	resetCmd.Flags().StringVarP(&resetWorkspace, "workspace", "w",
		"", "reset configuration of a specific workspace"+
			"(Kong Enterprise only).")
	resetCmd.Flags().BoolVar(&resetAllWorkspaces, "all-workspaces",
		false, "reset configuration of all workspaces (Kong Enterprise only).")
	resetCmd.Flags().StringSliceVar(&dumpConfig.SelectorTags,
		"select-tag", []string{},
		"only entities matching tags specified via this flag are deleted.\n"+
			"Multiple tags are ANDed together.")
	resetCmd.Flags().BoolVar(&dumpConfig.RBACResourcesOnly, "rbac-resources-only",
		false, "reset only the RBAC resources (Kong Enterprise only)")
}
