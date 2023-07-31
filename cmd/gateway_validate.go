package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/blang/semver/v4"
	"github.com/kong/deck/utils"
	"github.com/kong/deck/validate"
	"github.com/kong/go-database-reconciler/pkg/dump"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/state"
	reconcilerUtils "github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
	"github.com/spf13/cobra"
)

var (
	validateCmdKongStateFile     []string
	validateCmdRBACResourcesOnly bool
	validateOnline               bool
	validateWorkspace            string
	validateParallelism          int
	validateKonnectCompatibility bool
)

func executeValidate(cmd *cobra.Command, _ []string) error {
	mode := getMode(nil)
	if validateOnline && mode == modeKonnect {
		return fmt.Errorf("online validation not yet supported in konnect mode")
	}
	_ = sendAnalytics("validate", "", mode)
	// read target file
	// this does json schema validation as well
	targetContent, err := file.GetContentFromFiles(validateCmdKongStateFile, false)
	if err != nil {
		return err
	}

	dummyEmptyState, err := state.NewKongState()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var kongClient *kong.Client
	if validateOnline {
		kongClient, err = getKongClient(ctx, targetContent)
		if err != nil {
			return err
		}

		// if this is an online validation, we need to look up upstream consumers if required.
		lookUpSelectorTagsConsumers, err := determineLookUpSelectorTagsConsumers(*targetContent)
		if err != nil {
			return fmt.Errorf("error determining lookup selector tags for consumers: %w", err)
		}

		if lookUpSelectorTagsConsumers != nil {
			consumersGlobal, err := dump.GetAllConsumers(ctx, kongClient, lookUpSelectorTagsConsumers)
			if err != nil {
				return fmt.Errorf("error retrieving global consumers via lookup selector tags: %w", err)
			}
			for _, c := range consumersGlobal {
				targetContent.Consumers = append(targetContent.Consumers, file.FConsumer{Consumer: *c})
				if err != nil {
					return fmt.Errorf("error adding global consumer %v: %w", *c.Username, err)
				}
			}
		}

		// if this is an online validation, we need to look up upstream routes if required.
		lookUpSelectorTagsRoutes, err := determineLookUpSelectorTagsRoutes(*targetContent)
		if err != nil {
			return fmt.Errorf("error determining lookup selector tags for routes: %w", err)
		}

		if lookUpSelectorTagsRoutes != nil {
			routesGlobal, err := dump.GetAllRoutes(ctx, kongClient, lookUpSelectorTagsRoutes)
			if err != nil {
				return fmt.Errorf("error retrieving global routes via lookup selector tags: %w", err)
			}
			for _, r := range routesGlobal {
				targetContent.Routes = append(targetContent.Routes, file.FRoute{Route: *r})
				if err != nil {
					return fmt.Errorf("error adding global route %v: %w", r.FriendlyName(), err)
				}
			}
		}
	}

	rawState, err := file.Get(ctx, targetContent, file.RenderConfig{
		CurrentState: dummyEmptyState,
	}, dump.Config{}, kongClient)
	if err != nil {
		return err
	}
	if err := checkForRBACResources(*rawState, validateCmdRBACResourcesOnly); err != nil {
		return err
	}
	// this catches foreign relation errors
	ks, err := state.Get(rawState)
	if err != nil {
		return err
	}

	if validateKonnectCompatibility {
		if errs := validate.KonnectCompatibility(targetContent); len(errs) != 0 {
			return validate.ErrorsWrapper{Errors: errs}
		}
	}

	if validateOnline {
		if errs := validateWithKong(ctx, kongClient, ks, targetContent.FormatVersion); len(errs) != 0 {
			return validate.ErrorsWrapper{Errors: errs}
		}
	}
	return nil
}

// newValidateCmd represents the diff command
func newValidateCmd(deprecated bool, online bool) *cobra.Command {
	use := "validate [flags] [kong-state-files...]"
	short := "Validate the state file"
	long := `The validate command reads the state file and ensures validity.
It reads all the specified state files and reports YAML/JSON
parsing issues. It also checks for foreign relationships
and alerts if there are broken relationships, or missing links present.

`
	execute := executeValidate
	argsValidator := cobra.MinimumNArgs(0)
	var preRun func(cmd *cobra.Command, args []string) error

	if deprecated {
		use = "validate"
		short = "[deprecated] see 'deck [file|gateway] validate --help' for changes to the commands"
		long = `The validate command reads the state file and ensures validity.
It reads all the specified state files and reports YAML/JSON
parsing issues. It also checks for foreign relationships
and alerts if there are broken relationships, or missing links present.

No communication takes places between decK and Kong during the execution of
this command unless --online flag is used.
`

		execute = func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stderr, "Info: 'deck validate' functionality has moved to 'deck [file|gateway] validate' and\n"+
				"will be removed  in a future MAJOR version of deck.\n"+
				"Migration to 'deck [file|gateway] validate' is recommended.\n"+
				"   Note: - see 'deck [file|gateway] validate --help' for changes to the command\n"+
				"         - files changed to positional arguments without the '-s/--state' flag\n"+
				"         - the '--online' flag is removed, use either 'deck file' or 'deck gateway'\n"+
				"         - the default changed from 'kong.yaml' to '-' (stdin/stdout)\n")

			return executeValidate(cmd, args)
		}
		argsValidator = validateNoArgs
		preRun = func(_ *cobra.Command, _ []string) error {
			if len(diffCmdKongStateFile) == 0 {
				return fmt.Errorf("a state file with Kong's configuration " +
					"must be specified using `-s`/`--state` flag")
			}
			return preRunSilenceEventsFlag()
		}
	} else {
		preRun = func(_ *cobra.Command, args []string) error {
			validateOnline = online
			validateCmdKongStateFile = args
			if len(validateCmdKongStateFile) == 0 {
				validateCmdKongStateFile = []string{"-"}
			}
			return preRunSilenceEventsFlag()
		}

		if validateOnline {
			short = short + " (online)"
			long = long + "Validates against the Kong API, via communication with Kong. This increases the\n" +
				"time for validation but catches significant errors. No resource is created in Kong.\n" +
				"For offline validation see 'deck file validate'.\n"
		} else {
			short = short + " (locally)"
			long = long + "No communication takes places between decK and Kong during the execution of\n" +
				"this command. This is faster than the online validation, but catches fewer errors.\n" +
				"For online validation see 'deck gateway validate'.\n"
		}
	}

	validateCmd := &cobra.Command{
		Use:     use,
		Short:   short,
		Long:    long,
		Args:    argsValidator,
		RunE:    execute,
		PreRunE: preRun,
	}

	validateCmd.Flags().BoolVar(&validateCmdRBACResourcesOnly, "rbac-resources-only",
		false, "indicate that the state file(s) contains RBAC resources only (Kong Enterprise only).")
	if deprecated {
		validateCmd.Flags().StringSliceVarP(&validateCmdKongStateFile,
			"state", "s", []string{"kong.yaml"}, "file(s) containing Kong's configuration.\n"+
				"This flag can be specified multiple times for multiple files.\n"+
				"Use '-' to read from stdin.")
		validateCmd.Flags().BoolVar(&validateOnline, "online",
			false, "perform validations against Kong API. When this flag is used, validation is done\n"+
				"via communication with Kong. This increases the time for validation but catches \n"+
				"significant errors. No resource is created in Kong.")
	}
	validateCmd.Flags().StringVarP(&validateWorkspace, "workspace", "w",
		"", "validate configuration of a specific workspace "+
			"(Kong Enterprise only).\n"+
			"This takes precedence over _workspace fields in state files.")
	validateCmd.Flags().IntVar(&validateParallelism, "parallelism",
		10, "Maximum number of concurrent requests to Kong.")
	validateCmd.Flags().BoolVar(&validateKonnectCompatibility, "konnect-compatibility",
		false, "validate that the state file(s) are ready to be deployed to Konnect")

	if err := ensureGetAllMethods(); err != nil {
		panic(err.Error())
	}
	return validateCmd
}

func validateWithKong(
	ctx context.Context,
	kongClient *kong.Client,
	ks *state.KongState,
	formatVersion string,
) []error {
	if formatVersion == "" {
		formatVersion = utils.DefaultFormatVersion
	}
	parsedFormatVersion, err := semver.ParseTolerant(formatVersion)
	if err != nil {
		return []error{fmt.Errorf("parsing Kong version: %w", err)}
	}
	opts := validate.ValidatorOpts{
		Ctx:               ctx,
		State:             ks,
		Client:            kongClient,
		Parallelism:       validateParallelism,
		RBACResourcesOnly: validateCmdRBACResourcesOnly,
	}
	validator := validate.NewValidator(opts)
	return validator.Validate(parsedFormatVersion)
}

func getKongClient(ctx context.Context, targetContent *file.Content) (*kong.Client, error) {
	workspaceName := validateWorkspace
	if validateWorkspace != "" {
		// check if workspace exists
		workspaceName := getWorkspaceName(validateWorkspace, targetContent, false)
		workspaceExists, err := workspaceExists(ctx, rootConfig, workspaceName)
		if err != nil {
			return nil, err
		}
		if !workspaceExists {
			return nil, fmt.Errorf("workspace doesn't exist: %s", workspaceName)
		}
	}

	wsConfig := rootConfig.ForWorkspace(workspaceName)
	kongClient, err := reconcilerUtils.GetKongClient(wsConfig)
	if err != nil {
		return nil, err
	}
	return kongClient, nil
}

// ensureGetAllMethod ensures at init time that `GetAll()` method exists on the relevant structs.
// If the method doesn't exist, the code will panic. This increases the likelihood of catching such an
// error during manual testing.
func ensureGetAllMethods() error {
	// let's make sure ASAP that all resources have the expected GetAll method
	dummyEmptyState, _ := state.NewKongState()
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.Services); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.ACLGroups); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.BasicAuths); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.CACertificates); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.Certificates); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.Consumers); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.Documents); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.HMACAuths); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.JWTAuths); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.KeyAuths); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.Oauth2Creds); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.Plugins); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.Routes); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.SNIs); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.Targets); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.Upstreams); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.RBACEndpointPermissions); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.RBACRoles); err != nil {
		return err
	}
	if _, err := reconcilerUtils.CallGetAll(dummyEmptyState.FilterChains); err != nil {
		return err
	}
	return nil
}
