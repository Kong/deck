package cmd

import (
	"context"
	"strings"

	"github.com/kong/deck/dump"
	"github.com/kong/deck/file"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	dumpCmdKongStateFile string
	dumpCmdStateFormat   string
	dumpWorkspace        string
	dumpAllWorkspaces    bool
	dumpWithID           bool
)

func listWorkspaces(ctx context.Context, client *kong.Client) ([]string, error) {
	workspaces, err := client.Workspaces.ListAll(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fetching workspaces from Kong")
	}
	var res []string
	for _, workspace := range workspaces {
		res = append(res, *workspace.Name)
	}

	return res, nil
}

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Export Kong configuration to a file",
	Long: `Dump command reads all the entities present in Kong
and writes them to a file on disk.

The file can then be read using the Sync o Diff command to again
configure Kong.`,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		wsClient, err := utils.GetKongClient(rootConfig)
		if err != nil {
			return err
		}

		format := file.Format(strings.ToUpper(dumpCmdStateFormat))

		// Kong Enterprise dump all workspace
		if dumpAllWorkspaces {
			if dumpWorkspace != "" {
				return errors.New("workspace cannot be specified with --all-workspace flag")
			}
			if dumpCmdKongStateFile != "kong" {
				return errors.New("output-file cannot be specified with --all-workspace flag")
			}
			workspaces, err := listWorkspaces(cmd.Context(), wsClient)
			if err != nil {
				return err
			}

			for _, workspace := range workspaces {
				wsClient, err := utils.GetKongClient(rootConfig.ForWorkspace(workspace))
				if err != nil {
					return err
				}

				rawState, err := dump.Get(wsClient, dumpConfig)
				if err != nil {
					return errors.Wrap(err, "reading configuration from Kong")
				}
				ks, err := state.Get(rawState)
				if err != nil {
					return errors.Wrap(err, "building state")
				}

				if err := file.KongStateToFile(ks, file.WriteConfig{
					SelectTags: dumpConfig.SelectorTags,
					Workspace:  workspace,
					Filename:   workspace,
					FileFormat: format,
					WithID:     dumpWithID,
				}); err != nil {
					return err
				}
			}
			return nil
		}

		if yes, err := utils.ConfirmFileOverwrite(dumpCmdKongStateFile, dumpCmdStateFormat, assumeYes); err != nil {
			return err
		} else if !yes {
			return nil
		}

		// Kong OSS
		// or Kong Enterprise single workspace
		if dumpWorkspace != "" {
			wsConfig := rootConfig.ForWorkspace(dumpWorkspace)

			exists, err := workspaceExists(wsConfig)
			if err != nil {
				return err
			}
			if !exists {
				return errors.Errorf("workspace '%v' does not exist in Kong", dumpWorkspace)
			}

			wsClient, err = utils.GetKongClient(wsConfig)
			if err != nil {
				return err
			}
		}

		rawState, err := dump.Get(wsClient, dumpConfig)
		if err != nil {
			return errors.Wrap(err, "reading configuration from Kong")
		}
		ks, err := state.Get(rawState)
		if err != nil {
			return errors.Wrap(err, "building state")
		}
		if err := file.KongStateToFile(ks, file.WriteConfig{
			SelectTags: dumpConfig.SelectorTags,
			Workspace:  dumpWorkspace,
			Filename:   dumpCmdKongStateFile,
			FileFormat: format,
			WithID:     dumpWithID,
		}); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().StringVarP(&dumpCmdKongStateFile, "output-file", "o",
		"kong", "file to which to write Kong's configuration."+
			"Use '-' to write to stdout.")
	dumpCmd.Flags().StringVar(&dumpCmdStateFormat, "format",
		"yaml", "output file format: json or yaml")
	dumpCmd.Flags().BoolVar(&dumpWithID, "with-id",
		false, "write ID of all entities in the output")
	dumpCmd.Flags().StringVarP(&dumpWorkspace, "workspace", "w",
		"", "dump configuration of a specific workspace"+
			"(Kong Enterprise only).")
	dumpCmd.Flags().BoolVar(&dumpAllWorkspaces, "all-workspaces",
		false, "dump configuration of all workspaces (Kong Enterprise only).")
	dumpCmd.Flags().BoolVar(&dumpConfig.SkipConsumers, "skip-consumers",
		false, "skip exporting consumers and any plugins associated "+
			"with consumers")
	dumpCmd.Flags().StringSliceVar(&dumpConfig.SelectorTags,
		"select-tag", []string{},
		"only entities matching tags specified via this flag are exported.\n"+
			"Multiple tags are ANDed together.")
	dumpCmd.Flags().BoolVar(&dumpConfig.RBACResourcesOnly, "rbac-resources-only",
		false, "export only the RBAC resources (Kong Enterprise only)")
	dumpCmd.Flags().BoolVar(&assumeYes, "yes",
		false, "Assume 'yes' to prompts and run non-interactively")
}
