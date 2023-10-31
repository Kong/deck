package cmd

import (
	"context"
	"fmt"

	"github.com/kong/deck/cprint"
	"github.com/kong/deck/dump"
	"github.com/kong/deck/file"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/deck/validate"
	"github.com/kong/go-kong/kong"
	"github.com/spf13/cobra"
)

var (
	validateCmdKongStateFile     []string
	validateCmdRBACResourcesOnly bool
	validateOnline               bool
	validateWorkspace            string
	validateParallelism          int
	validateJSONOutput           bool
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

	if validateOnline {
		if errs := validateWithKong(ctx, kongClient, ks); len(errs) != 0 {
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
	preRun := func(cmd *cobra.Command, args []string) error {
		diffCmdKongStateFile = args
		if len(diffCmdKongStateFile) == 0 {
			diffCmdKongStateFile = []string{"-"}
		}
		return preRunSilenceEventsFlag()
	}

	if deprecated {
		use = "validate"
		short = "[deprecated] use 'gateway validate' or 'file validate' instead"
		long = `The validate command reads the state file and ensures validity.
It reads all the specified state files and reports YAML/JSON
parsing issues. It also checks for foreign relationships
and alerts if there are broken relationships, or missing links present.

No communication takes places between decK and Kong during the execution of
this command unless --online flag is used.
`

		execute = func(cmd *cobra.Command, args []string) error {
			cprint.UpdatePrintf("Warning: 'deck validate' is DEPRECATED and will be removed in a future version. " +
				"Use 'deck gateway validate' instead.\n")
			return executeValidate(cmd, args)
		}
		argsValidator = validateNoArgs
		preRun = func(cmd *cobra.Command, args []string) error {
			if len(diffCmdKongStateFile) == 0 {
				return fmt.Errorf("a state file with Kong's configuration " +
					"must be specified using `-s`/`--state` flag")
			}
			return preRunSilenceEventsFlag()
		}
	} else {
		validateOnline = online
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
	validateCmd.Flags().BoolVar(&validateJSONOutput, "json-output",
		false, "generate command execution report in a JSON format")

	if err := ensureGetAllMethods(); err != nil {
		panic(err.Error())
	}
	return validateCmd
}

func validateWithKong(ctx context.Context, kongClient *kong.Client, ks *state.KongState) []error {
	// make sure we are able to connect to Kong
	kongVersion, err := fetchKongVersion(ctx, rootConfig)
	if err != nil {
		return []error{fmt.Errorf("couldn't fetch Kong version: %w", err)}
	}
	parsedKongVersion, err := utils.ParseKongVersion(kongVersion)
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
	return validator.Validate(parsedKongVersion)
}

func getKongClient(ctx context.Context, targetContent *file.Content) (*kong.Client, error) {
	workspaceName := validateWorkspace
	if validateWorkspace != "" {
		// check if workspace exists
		workspaceName := getWorkspaceName(validateWorkspace, targetContent, validateJSONOutput)
		workspaceExists, err := workspaceExists(ctx, rootConfig, workspaceName)
		if err != nil {
			return nil, err
		}
		if !workspaceExists {
			return nil, fmt.Errorf("workspace doesn't exist: %s", workspaceName)
		}
	}

	wsConfig := rootConfig.ForWorkspace(workspaceName)
	kongClient, err := utils.GetKongClient(wsConfig)
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
	if _, err := utils.CallGetAll(dummyEmptyState.Services); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.ACLGroups); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.BasicAuths); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.CACertificates); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.Certificates); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.Consumers); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.Documents); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.HMACAuths); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.JWTAuths); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.KeyAuths); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.Oauth2Creds); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.Plugins); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.Routes); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.SNIs); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.Targets); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.Upstreams); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.RBACEndpointPermissions); err != nil {
		return err
	}
	if _, err := utils.CallGetAll(dummyEmptyState.RBACRoles); err != nil {
		return err
	}
	return nil
}
