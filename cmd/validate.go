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
	validateOnlineErrors         []error
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
			validateWithKong(cmd, ks)
			for _, e := range validateOnlineErrors {
				cprint.DeletePrintln(e)
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

func validateEntities(ctx context.Context, obj interface{}, kongClient *kong.Client, entityType string) {
	// call GetAll on entity
	method := reflect.ValueOf(obj).MethodByName("GetAll")
	entities := method.Call([]reflect.Value{})[0].Interface()
	values := reflect.ValueOf(entities)
	var wg sync.WaitGroup
	wg.Add(values.Len())
	for i := 0; i < values.Len(); i++ {
		go func(i int) {
			defer wg.Done()
			if err := validate(ctx, values.Index(i).Interface(), kongClient, entityType); err != nil {
				validateOnlineErrors = append(validateOnlineErrors, err)
			}
		}(i)
	}
	wg.Wait()
}

func validateWithKong(cmd *cobra.Command, ks *state.KongState) error {
	ctx := cmd.Context()
	kongClient, err := utils.GetKongClient(rootConfig)
	if err != nil {
		return err
	}
	validateEntities(ctx, ks.Services, kongClient, "services")
	validateEntities(ctx, ks.ACLGroups, kongClient, "acls")
	validateEntities(ctx, ks.BasicAuths, kongClient, "basicauth_credentials")
	validateEntities(ctx, ks.CACertificates, kongClient, "ca_certificates")
	validateEntities(ctx, ks.Certificates, kongClient, "certificates")
	validateEntities(ctx, ks.Consumers, kongClient, "consumers")
	validateEntities(ctx, ks.Documents, kongClient, "documents")
	validateEntities(ctx, ks.HMACAuths, kongClient, "hmacauth_credentials")
	validateEntities(ctx, ks.JWTAuths, kongClient, "jwt_secrets")
	validateEntities(ctx, ks.KeyAuths, kongClient, "keyauth_credentials")
	validateEntities(ctx, ks.Oauth2Creds, kongClient, "oauth2_credentials")
	validateEntities(ctx, ks.Plugins, kongClient, "plugins")
	validateEntities(ctx, ks.Routes, kongClient, "routes")
	validateEntities(ctx, ks.SNIs, kongClient, "snis")
	validateEntities(ctx, ks.Targets, kongClient, "targets")
	validateEntities(ctx, ks.Upstreams, kongClient, "upstreams")
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
		false, "perform schema validation against Kong API.")
}
