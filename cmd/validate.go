package cmd

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/kong/deck/cprint"
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
)

// validateCmd represents the diff command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the state file",
	Long: `The validate command reads the state file and ensures validity.
It reads all the specified state files and reports YAML/JSON
parsing issues. It also checks for foreign relationships
and alerts if there are broken relationships, or missing links present.

When ran with the --online flag, it also uses the 'validate' endpoint in Kong
to validate entities configuration.
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
			if err := validateWithKong(cmd, ks); err != nil {
				for _, e := range err {
					cprint.DeletePrintln(e)
				}
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

func validate(ctx context.Context, entity interface{}, kongClient *kong.Client, entityType string) error {
	_, err := validateEntity(ctx, kongClient, entityType, entity)
	if err != nil {
		return err
	}
	return nil
}

func validateEntities(ctx context.Context, obj interface{}, kongClient *kong.Client, entityType string) []error {
	entities := callGetAll(obj)
	errors := []error{}
	var wg sync.WaitGroup
	wg.Add(entities.Len())
	mu := &sync.Mutex{}
	for i := 0; i < entities.Len(); i++ {
		go func(i int) {
			defer wg.Done()
			if err := validate(ctx, entities.Index(i).Interface(), kongClient, entityType); err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return errors
}

func validateWithKong(cmd *cobra.Command, ks *state.KongState) []error {
	ctx := cmd.Context()
	kongClient, err := utils.GetKongClient(rootConfig)
	allErr := []error{}
	if err != nil {
		return []error{err}
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

func callGetAll(obj interface{}) reflect.Value {
	// call GetAll method on entity
	method := reflect.ValueOf(obj).MethodByName("GetAll")
	entities := method.Call([]reflect.Value{})[0].Interface()
	return reflect.ValueOf(entities)
}

func validateGetAllMethod() {
	dummyEmptyState, _ := state.NewKongState()
	callGetAll(dummyEmptyState.Services)
	callGetAll(dummyEmptyState.ACLGroups)
	callGetAll(dummyEmptyState.BasicAuths)
	callGetAll(dummyEmptyState.CACertificates)
	callGetAll(dummyEmptyState.Certificates)
	callGetAll(dummyEmptyState.Consumers)
	callGetAll(dummyEmptyState.Documents)
	callGetAll(dummyEmptyState.HMACAuths)
	callGetAll(dummyEmptyState.JWTAuths)
	callGetAll(dummyEmptyState.KeyAuths)
	callGetAll(dummyEmptyState.Oauth2Creds)
	callGetAll(dummyEmptyState.Plugins)
	callGetAll(dummyEmptyState.Routes)
	callGetAll(dummyEmptyState.SNIs)
	callGetAll(dummyEmptyState.Targets)
	callGetAll(dummyEmptyState.Upstreams)
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
		false, "perform schema validation against Kong API.")
	validateGetAllMethod()
}
