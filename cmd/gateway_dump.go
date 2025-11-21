package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kong/deck/sanitize"
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
	sanitizationSalt               string
)

func listWorkspaces(ctx context.Context, client *kong.Client) ([]string, error) {
	workspaces, err := client.Workspaces.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching workspaces from Kong: %w", err)
	}
	res := make([]string, 0, len(workspaces))
	for _, workspace := range workspaces {
		res = append(res, *workspace.Name)
	}

	return res, nil
}

func getWorkspaceClient(ctx context.Context, workspace string) (*kong.Client, error) {
	wsConfig := rootConfig.ForWorkspace(workspace)
	exists, err := workspaceExists(ctx, rootConfig, workspace)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("workspace '%v' does not exist in Kong", workspace)
	}

	wsClient, err := utils.GetKongClient(wsConfig)
	if err != nil {
		return nil, err
	}

	return wsClient, nil
}

func getKongState(ctx context.Context, wsClient *kong.Client) (*state.KongState, error) {
	rawState, err := dump.Get(ctx, wsClient, dumpConfig)
	if err != nil {
		return nil, fmt.Errorf("reading configuration from Kong: %w", err)
	}
	ks, err := state.Get(rawState)
	if err != nil {
		return nil, fmt.Errorf("building state: %w", err)
	}
	return ks, nil
}

func sanitizeContent(ctx context.Context, client *kong.Client,
	ks *state.KongState, writeConfig file.WriteConfig, isKonnect bool,
) error {
	writeConfig.WithID = true // always write IDs for sanitization
	fileContent, err := file.KongStateToContent(ks, writeConfig)
	if err != nil {
		return fmt.Errorf("sanitizing content: %w", err)
	}

	sanitizer := sanitize.NewSanitizer(&sanitize.SanitizerOptions{
		Ctx:       ctx,
		Client:    client,
		Content:   fileContent,
		IsKonnect: isKonnect,
		Salt:      sanitizationSalt,
	})

	sanitizedContent, err := sanitizer.Sanitize()
	if err != nil {
		return fmt.Errorf("sanitizing content: %w", err)
	}

	return file.WriteContentToFile(sanitizedContent, writeConfig.Filename, writeConfig.FileFormat)
}

func executeDump(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	if yes, err := utils.ConfirmFileOverwrite(dumpCmdKongStateFile, dumpCmdStateFormat, assumeYes); err != nil {
		return err
	} else if !yes {
		return nil
	}

	if inKonnectMode(nil) {
		// Konnect ConsumerGroup APIs don't support the query-parameter list_consumers yet
		if dumpConfig.SkipConsumersWithConsumerGroups {
			return errors.New("the flag --skip-consumers-with-consumer-groups can not be used with Konnect")
		}

		_ = sendAnalytics("dump", "", modeKonnect)
		return dumpKonnectV2(ctx)
	}

	wsClient, err := utils.GetKongClient(rootConfig)
	if err != nil {
		return err
	}

	format := file.Format(strings.ToUpper(dumpCmdStateFormat))

	kongVersion, err := fetchKongVersion(ctx, rootConfig.ForWorkspace(dumpWorkspace))
	if err != nil {
		return fmt.Errorf("reading Kong version: %w", err)
	}
	_ = sendAnalytics("dump", kongVersion, modeKong)

	writeConfig := file.WriteConfig{
		SelectTags:                       dumpConfig.SelectorTags,
		Workspace:                        dumpWorkspace,
		Filename:                         dumpCmdKongStateFile,
		FileFormat:                       format,
		WithID:                           dumpWithID,
		KongVersion:                      kongVersion,
		IsConsumerGroupPolicyOverrideSet: dumpConfig.IsConsumerGroupPolicyOverrideSet,
		SanitizeContent:                  dumpConfig.SanitizeContent,
	}

	// Kong Enterprise dump all workspace
	if dumpAllWorkspaces {
		workspaces, err := listWorkspaces(ctx, wsClient)
		if err != nil {
			return err
		}

		for _, workspace := range workspaces {
			wsClient, err = getWorkspaceClient(ctx, workspace)
			if err != nil {
				return fmt.Errorf("getting Kong client for workspace '%s': %w", workspace, err)
			}

			ks, err := getKongState(ctx, wsClient)
			if err != nil {
				return fmt.Errorf("getting Kong state for workspace '%s': %w", workspace, err)
			}

			writeConfig.Workspace = workspace
			writeConfig.Filename = workspace

			if dumpConfig.SanitizeContent {
				if err := sanitizeContent(ctx, wsClient, ks, writeConfig, false); err != nil {
					return fmt.Errorf("sanitizing content for workspace '%s': %w", workspace, err)
				}
			} else {
				if err := file.KongStateToFile(ks, writeConfig); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Kong OSS
	// or Kong Enterprise single workspace
	if dumpWorkspace != "" {
		wsClient, err = getWorkspaceClient(ctx, dumpWorkspace)
		if err != nil {
			return err
		}
	}

	ks, err := getKongState(ctx, wsClient)
	if err != nil {
		return fmt.Errorf("getting Kong state: %w", err)
	}

	if dumpConfig.SanitizeContent {
		return sanitizeContent(ctx, wsClient, ks, writeConfig, false)
	}

	return file.KongStateToFile(ks, writeConfig)
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
	dumpCmd.Flags().BoolVar(&dumpConfig.SkipConsumersWithConsumerGroups, "skip-consumers-with-consumer-groups",
		false, "do not show the association between consumer and consumer-group.\n"+
			"If set to true, deck skips listing consumers with consumer-groups,\n"+
			"thus gaining some performance with large configs. This flag is not valid with Konnect.")
	dumpCmd.Flags().BoolVar(&dumpConfig.IsConsumerGroupPolicyOverrideSet, "consumer-group-policy-overrides",
		false, "allow deck to dump consumer-group policy overrides.\n"+
			"This allows policy overrides to work with Kong GW versions >= 3.4\n"+
			"Warning: do not mix with consumer-group scoped plugins")
	dumpCmd.Flags().BoolVar(&dumpConfig.SkipDefaults, "skip-defaults",
		false, "skip exporting default values.")
	dumpCmd.Flags().BoolVar(&dumpConfig.SanitizeContent, "sanitize",
		false, "dumps a sanitized version of the gateway configuration.\n"+
			"This feature hashes passwords, keys and other sensitive details.")
	dumpCmd.Flags().StringVar(&sanitizationSalt, "sanitization-salt",
		"", "salt used to hash sensitive data in the sanitized dump.\n"+
			"Use this flag to ensure that the same sensitive data is hashed to the same value.\n"+
			"If not set, a random salt is used.\n")
	// This flag is hidden for now. We can mark it as visible in the future
	// if we decide to expose it to the user. For now, it is used internally
	// for testing purposes.
	// Discarding the error as it is not critical
	_ = dumpCmd.Flags().MarkHidden("sanitization-salt")
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
