package cmd

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/kong/deck/file"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/spf13/cobra"
)

var (
	validateCmdKongStateFile     []string
	validateCmdRBACResourcesOnly bool
	validateOnline               bool
	validateWorkspace            string
	validateParallelism          int
)

type validateErrorsWrapper struct {
	errors []error
}

func (v validateErrorsWrapper) Error() string {
	var errStr string
	for _, e := range v.errors {
		errStr += fmt.Sprintf("%s\n", e.Error())
	}
	return errStr
}

// validateCmd represents the diff command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the state file",
	Long: `The validate command reads the state file and ensures validity.
It reads all the specified state files and reports YAML/JSON
parsing issues. It also checks for foreign relationships
and alerts if there are broken relationships, or missing links present.

No communication takes places between decK and Kong during the execution of
this command unless --online flag is used.
`,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = sendAnalytics("validate", "")
		// read target file
		// this does json schema validation as well
		targetContent, err := file.GetContentFromFiles(validateCmdKongStateFile)
		if err != nil {
			return err
		}

		dummyEmptyState, err := state.NewKongState()
		if err != nil {
			return err
		}

		rawState, err := file.Get(targetContent, file.RenderConfig{
			CurrentState: dummyEmptyState,
		})
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
			if errs := validateWithKong(cmd, ks, targetContent); len(errs) != 0 {
				return validateErrorsWrapper{errors: errs}
			}
		}
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(validateCmdKongStateFile) == 0 {
			return fmt.Errorf("a state file with Kong's configuration " +
				"must be specified using -s/--state flag")
		}
		return nil
	},
}

func validateEntities(ctx context.Context, obj interface{}, kongClient *kong.Client, entityType string) []error {
	entities, err := callGetAll(obj)
	if err != nil {
		return []error{err}
	}
	errors := []error{}

	// create a buffer of channels. Creation of new coroutines
	// are allowed only if the buffer is not full.
	chanBuff := make(chan struct{}, validateParallelism)

	var wg sync.WaitGroup
	wg.Add(entities.Len())
	// each coroutine will append on a slice of errors.
	// since slices are not thread-safe, let's add a mutex
	// to handle access to the slice.
	mu := &sync.Mutex{}
	for i := 0; i < entities.Len(); i++ {
		// reserve a slot
		chanBuff <- struct{}{}
		go func(i int) {
			defer wg.Done()
			// release a slot when completed
			defer func() { <-chanBuff }()
			_, err := validateEntity(ctx, kongClient, entityType, entities.Index(i).Interface())
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return errors
}

func validateWithKong(cmd *cobra.Command, ks *state.KongState, targetContent *file.Content) []error {
	ctx := cmd.Context()
	// make sure we are able to connect to Kong
	_, err := fetchKongVersion(ctx, rootConfig)
	if err != nil {
		return []error{fmt.Errorf("couldn't fetch Kong version: %w", err)}
	}

	// check if workspace exists
	workspaceName := getWorkspaceName(validateWorkspace, targetContent)
	workspaceExists, err := workspaceExists(ctx, rootConfig, workspaceName)
	if err != nil {
		return []error{err}
	}
	if !workspaceExists {
		return []error{fmt.Errorf("workspace doesn't exist: %s", workspaceName)}
	}

	wsConfig := rootConfig.ForWorkspace(workspaceName)
	kongClient, err := utils.GetKongClient(wsConfig)
	allErr := []error{}
	if err != nil {
		return []error{err}
	}

	// validate RBAC resources first.
	if err := validateEntities(ctx, ks.RBACEndpointPermissions, kongClient, "rbac-endpointpermission"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.RBACRoles, kongClient, "rbac-role"); err != nil {
		allErr = append(allErr, err...)
	}
	if validateCmdRBACResourcesOnly {
		return allErr
	}

	if err := validateEntities(ctx, ks.Services, kongClient, "services"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.ACLGroups, kongClient, "acls"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.BasicAuths, kongClient, "basicauth_credentials"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.CACertificates, kongClient, "ca_certificates"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.Certificates, kongClient, "certificates"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.Consumers, kongClient, "consumers"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.Documents, kongClient, "documents"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.HMACAuths, kongClient, "hmacauth_credentials"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.JWTAuths, kongClient, "jwt_secrets"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.KeyAuths, kongClient, "keyauth_credentials"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.Oauth2Creds, kongClient, "oauth2_credentials"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.Plugins, kongClient, "plugins"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.Routes, kongClient, "routes"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.SNIs, kongClient, "snis"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.Targets, kongClient, "targets"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := validateEntities(ctx, ks.Upstreams, kongClient, "upstreams"); err != nil {
		allErr = append(allErr, err...)
	}
	return allErr
}

func callGetAll(obj interface{}) (reflect.Value, error) {
	// call GetAll method on entity
	var result reflect.Value
	method := reflect.ValueOf(obj).MethodByName("GetAll")
	if !method.IsValid() {
		return result, fmt.Errorf("GetAll() method not found for %v. "+
			"Please file a bug with Kong Inc.", reflect.ValueOf(obj).Type())
	}
	entities := method.Call([]reflect.Value{})[0].Interface()
	result = reflect.ValueOf(entities)
	return result, nil
}

// ensureGetAllMethod ensures at init time that `GetAll()` method exists on the relevant structs.
// If the method doesn't exist, the code will panic. This increases the likelihood of catching such an
// error during manual testing.
func ensureGetAllMethods() error {
	// let's make sure ASAP that all resources have the expected GetAll method
	dummyEmptyState, _ := state.NewKongState()
	if _, err := callGetAll(dummyEmptyState.Services); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.ACLGroups); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.BasicAuths); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.CACertificates); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.Certificates); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.Consumers); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.Documents); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.HMACAuths); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.JWTAuths); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.KeyAuths); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.Oauth2Creds); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.Plugins); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.Routes); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.SNIs); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.Targets); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.Upstreams); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.RBACEndpointPermissions); err != nil {
		return err
	}
	if _, err := callGetAll(dummyEmptyState.RBACRoles); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVar(&validateCmdRBACResourcesOnly, "rbac-resources-only",
		false, "indicate that the state file(s) contains RBAC resources only (Kong Enterprise only).")
	validateCmd.Flags().StringSliceVarP(&validateCmdKongStateFile,
		"state", "s", []string{"kong.yaml"}, "file(s) containing Kong's configuration.\n"+
			"This flag can be specified multiple times for multiple files.\n"+
			"Use '-' to read from stdin.")
	validateCmd.Flags().BoolVar(&validateOnline, "online",
		false, "perform validations against Kong API. When this flag is used, validation is done\n"+
			"via communication with Kong. This increases the time for validation but catches \n"+
			"significant errors. No resource is created in Kong.")
	validateCmd.Flags().StringVarP(&validateWorkspace, "workspace", "w",
		"", "validate configuration of a specific workspace "+
			"(Kong Enterprise only).\n"+
			"This takes precedence over _workspace fields in state files.")
	validateCmd.Flags().IntVar(&validateParallelism, "parallelism",
		10, "Maximum number of concurrent requests to Kong.")
	if err := ensureGetAllMethods(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
