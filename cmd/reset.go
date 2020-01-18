package cmd

import (
	"fmt"
	"strings"

	"github.com/hbagdi/deck/dump"
	"github.com/hbagdi/deck/reset"
	"github.com/hbagdi/deck/utils"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		if !resetCmdForce {
			ok, err := confirm()
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		client, err := utils.GetKongClient(config)
		if err != nil {
			return err
		}
		// Kong OSS or default workspace
		if !resetAllWorkspaces && resetWorkspace == "" {
			state, err := dump.Get(client, dumpConfig)
			if err != nil {
				return err
			}
			err = reset.Reset(state, client)
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
			workspaces, err = listWorkspaces(client, config.Address)
			if err != nil {
				return err
			}
		}
		if resetWorkspace != "" {
			workspaces = append(workspaces, resetWorkspace)
		}

		for _, workspace := range workspaces {
			config.Workspace = workspace
			client, err := utils.GetKongClient(config)
			if err != nil {
				return err
			}
			state, err := dump.Get(client, dumpConfig)
			if err != nil {
				return err
			}
			err = reset.Reset(state, client)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

// confirm prompts a user for a confirmation
// and returns true with no error if input is "yes" or "y" (case-insensitive),
// otherwise false.
func confirm() (bool, error) {
	fmt.Println("This will delete all configuration from Kong's database.")
	fmt.Print("> Are you sure? ")
	yes := []string{"yes", "y"}
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return false, err
	}
	input = strings.ToLower(input)
	for _, valid := range yes {
		if input == valid {
			return true, nil
		}
	}
	return false, nil
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
}
