package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/dump"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/state"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
	"github.com/spf13/cobra"
)

const defaultFileOutName = "kong"

var (
	dumpCmdKongStateFileDeprecated string
	dumpCmdKongStateFile           string
	dumpCmdStateFormat             string
	dumpWorkspace                  string
	dumpAllWorkspaces              bool
	dumpWithID                     bool
)

func listWorkspaces(ctx context.Context, client *kong.Client) ([]string, error) {
	workspaces, err := client.Workspaces.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching workspaces from Kong: %w", err)
	}
	var res []string
	for _, workspace := range workspaces {
		res = append(res, *workspace.Name)
	}

	return res, nil
}

func executeDump(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	if yes, err := utils.ConfirmFileOverwrite(dumpCmdKongStateFile, dumpCmdStateFormat, assumeYes); err != nil {
		return err
	} else if !yes {
		return nil
	}

	if inKonnectMode(nil) {
		_ = sendAnalytics("dump", "", modeKonnect)
		return dumpKonnectV2(ctx)
	}

	wsClient, err := utils.GetKongClient(rootConfig)
	if err != nil {
		return err
	}

	dumpConfig.IsFilterChainsSupported, err = determineFilterChainSupport(ctx, wsClient)
	if err != nil {
		return err
	}

	format := file.Format(strings.ToUpper(dumpCmdStateFormat))

	kongVersion, err := fetchKongVersion(ctx, rootConfig.ForWorkspace(dumpWorkspace))
	if err != nil {
		return fmt.Errorf("reading Kong version: %w", err)
	}
	_ = sendAnalytics("dump", kongVersion, modeKong)

	// Kong Enterprise dump all workspace
	if dumpAllWorkspaces {
		workspaces, err := listWorkspaces(ctx, wsClient)
		if err != nil {
			return err
		}

		for _, workspace := range workspaces {
			wsClient, err := utils.GetKongClient(rootConfig.ForWorkspace(workspace))
			if err != nil {
				return err
			}

			rawState, err := dump.Get(ctx, wsClient, dumpConfig)
			if err != nil {
				return fmt.Errorf("reading configuration from Kong: %w", err)
			}
			ks, err := state.Get(rawState)
			if err != nil {
				return fmt.Errorf("building state: %w", err)
			}

			if err := file.KongStateToFile(ks, file.WriteConfig{
				SelectTags:  dumpConfig.SelectorTags,
				Workspace:   workspace,
				Filename:    workspace,
				FileFormat:  format,
				WithID:      dumpWithID,
				KongVersion: kongVersion,
			}); err != nil {
				return err
			}
		}
		return nil
	}

	// Kong OSS
	// or Kong Enterprise single workspace
	if dumpWorkspace != "" {
		wsConfig := rootConfig.ForWorkspace(dumpWorkspace)
		exists, err := workspaceExists(ctx, rootConfig, dumpWorkspace)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("workspace '%v' does not exist in Kong", dumpWorkspace)
		}
		wsClient, err = utils.GetKongClient(wsConfig)
		if err != nil {
			return err
		}
	}

	rawState, err := dump.Get(ctx, wsClient, dumpConfig)
	if err != nil {
		return fmt.Errorf("reading configuration from Kong: %w", err)
	}
	ks, err := state.Get(rawState)
	if err != nil {
		return fmt.Errorf("building state: %w", err)
	}
	return file.KongStateToFile(ks, file.WriteConfig{
		SelectTags:  dumpConfig.SelectorTags,
		Workspace:   dumpWorkspace,
		Filename:    dumpCmdKongStateFile,
		FileFormat:  format,
		WithID:      dumpWithID,
		KongVersion: kongVersion,
	})
}

// newDumpCmd represents the dump command
func newDumpCmd(deprecated bool) *cobra.Command {
	short := "Export Kong configuration to a file"
	execute := executeDump
	fileOutDefault := "-"
	if deprecated {
		short = "[deprecated] see 'deck gateway dump --help' for changes to the command"
		execute = func(cmd *cobra.Command, args []string) error {
			dumpCmdKongStateFile = dumpCmdKongStateFileDeprecated
			fmt.Fprintf(os.Stderr, "Info: 'deck dump' functionality has moved to 'deck gateway dump' and will be removed\n"+
				"in a future MAJOR version of deck. Migration to 'deck gateway dump' is recommended.\n"+
				"   Note: - see 'deck gateway dump --help' for changes to the command\n"+
				"         - the default changed from 'kong.yaml' to '-' (stdin/stdout)\n")
			return executeDump(cmd, args)
		}
		fileOutDefault = defaultFileOutName
	}

	dumpCmd := &cobra.Command{
		Use:   "dump",
		Short: short,
		Long: `The dump command reads all entities present in Kong
and writes them to a local file.

The file can then be read using the sync command or diff command to
configure Kong.`,
		Args: validateNoArgs,
		RunE: execute,
	}

	dumpCmd.Flags().StringVar(&dumpCmdStateFormat, "format",
		"yaml", "output file format: json or yaml.")
	dumpCmd.Flags().BoolVar(&dumpWithID, "with-id",
		false, "write ID of all entities in the output")
	dumpCmd.Flags().StringVarP(&dumpWorkspace, "workspace", "w",
		"", "dump configuration of a specific Workspace"+
			"(Kong Enterprise only).")
	dumpCmd.Flags().BoolVar(&dumpAllWorkspaces, "all-workspaces",
		false, "dump configuration of all Workspaces (Kong Enterprise only).")
	dumpCmd.Flags().BoolVar(&dumpConfig.SkipConsumers, "skip-consumers",
		false, "skip exporting consumers, consumer-groups and any plugins associated "+
			"with them.")
	dumpCmd.Flags().StringSliceVar(&dumpConfig.SelectorTags,
		"select-tag", []string{},
		"only entities matching tags specified with this flag are exported.\n"+
			"When this setting has multiple tag values, entities must match every tag.")
	dumpCmd.Flags().BoolVar(&dumpConfig.RBACResourcesOnly, "rbac-resources-only",
		false, "export only the RBAC resources (Kong Enterprise only).")
	dumpCmd.Flags().BoolVar(&assumeYes, "yes",
		false, "assume `yes` to prompts and run non-interactively.")
	dumpCmd.Flags().BoolVar(&dumpConfig.SkipCACerts, "skip-ca-certificates",
		false, "do not dump CA certificates.")
	if deprecated {
		dumpCmd.Flags().StringVarP(&dumpCmdKongStateFileDeprecated, "output-file", "o",
			fileOutDefault, "file to which to write Kong's configuration."+
				"Use `-` to write to stdout.")
	} else {
		dumpCmd.Flags().StringVarP(&dumpCmdKongStateFile, "output-file", "o",
			fileOutDefault, "file to which to write Kong's configuration."+
				"Use `-` to write to stdout.")
	}
	dumpCmd.MarkFlagsMutuallyExclusive("output-file", "all-workspaces")
	dumpCmd.MarkFlagsMutuallyExclusive("workspace", "all-workspaces")

	return dumpCmd
}
