//nolint:deadcode
package integration

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/kong/deck/cmd"
	deckDump "github.com/kong/deck/dump"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

func int32p(i int) *int32 {
	p := int32(i)
	return &p
}

func getKongAddress() string {
	address := os.Getenv("DECK_KONG_ADDR")
	if address != "" {
		return address
	}
	return "http://localhost:8001"
}

func getTestClient() (*kong.Client, error) {
	ctx := context.Background()
	konnectConfig := utils.KonnectConfig{
		Address:  os.Getenv("DECK_KONNECT_ADDR"),
		Email:    os.Getenv("DECK_KONNECT_EMAIL"),
		Password: os.Getenv("DECK_KONNECT_PASSWORD"),
	}
	if konnectConfig.Email != "" && konnectConfig.Password != "" {
		return cmd.GetKongClientForKonnectMode(ctx, &konnectConfig)
	}
	return utils.GetKongClient(utils.KongClientConfig{
		Address: getKongAddress(),
	})
}

func runWhenKonnect(t *testing.T) {
	if os.Getenv("DECK_KONNECT_EMAIL") == "" || os.Getenv("DECK_KONNECT_PASSWORD") == "" {
		t.Log("non-Konnect test instance, skipping")
		t.Skip()
	}
}

func skipWhenKonnect(t *testing.T) {
	if os.Getenv("DECK_KONNECT_EMAIL") != "" || os.Getenv("DECK_KONNECT_PASSWORD") != "" {
		t.Log("non-Kong test instance, skipping")
		t.Skip()
	}
}

func runWhen(t *testing.T, mode string, semverRange string) {
	switch mode {
	case "kong":
		skipWhenKonnect(t)
		kong.RunWhenKong(t, semverRange)
	case "enterprise":
		skipWhenKonnect(t)
		kong.RunWhenEnterprise(t, semverRange, kong.RequiredFeatures{})
	case "konnect":
		runWhenKonnect(t)
	}
}

func sortSlices(x, y interface{}) bool {
	var xName, yName string
	switch xEntity := x.(type) {
	case *kong.Service:
		yEntity := y.(*kong.Service)
		xName = *xEntity.Name
		yName = *yEntity.Name
	case *kong.Route:
		yEntity := y.(*kong.Route)
		xName = *xEntity.Name
		yName = *yEntity.Name
	case *kong.Vault:
		yEntity := y.(*kong.Vault)
		xName = *xEntity.Prefix
		yName = *yEntity.Prefix
	case *kong.Consumer:
		yEntity := y.(*kong.Consumer)
		xName = *xEntity.Username
		yName = *yEntity.Username
	case *kong.ConsumerGroup:
		yEntity := y.(*kong.ConsumerGroup)
		xName = *xEntity.Name
		yName = *yEntity.Name
	case *kong.ConsumerGroupObject:
		yEntity := y.(*kong.ConsumerGroupObject)
		xName = *xEntity.ConsumerGroup.Name
		yName = *yEntity.ConsumerGroup.Name
	case *kong.ConsumerGroupPlugin:
		yEntity := y.(*kong.ConsumerGroupPlugin)
		xName = *xEntity.ConsumerGroup.ID
		yName = *yEntity.ConsumerGroup.ID
	case *kong.KeyAuth:
		yEntity := y.(*kong.KeyAuth)
		xName = *xEntity.Key
		yName = *yEntity.Key
	case *kong.Plugin:
		yEntity := y.(*kong.Plugin)
		xName = *xEntity.Name
		yName = *yEntity.Name
		if xEntity.Route != nil {
			xName += *xEntity.Route.ID
		}
		if xEntity.Service != nil {
			xName += *xEntity.Service.ID
		}
		if xEntity.Consumer != nil {
			xName += *xEntity.Consumer.ID
		}
		if xEntity.Consumer != nil {
			xName += *xEntity.Consumer.ID
		}
		if yEntity.Route != nil {
			yName += *yEntity.Route.ID
		}
		if yEntity.Service != nil {
			yName += *yEntity.Service.ID
		}
		if yEntity.Consumer != nil {
			yName += *yEntity.Consumer.ID
		}
	}
	return xName < yName
}

func testKongState(t *testing.T, client *kong.Client, isKonnect bool,
	expectedState utils.KongRawState, ignoreFields []cmp.Option,
) {
	// Get entities from Kong
	ctx := context.Background()
	dumpConfig := deckDump.Config{}
	if expectedState.RBACEndpointPermissions != nil {
		dumpConfig.RBACResourcesOnly = true
	}
	if isKonnect {
		// use default RG for testing
		dumpConfig.KonnectRuntimeGroup = "default"
	}
	kongState, err := deckDump.Get(ctx, client, dumpConfig)
	if err != nil {
		t.Errorf(err.Error())
	}

	opt := []cmp.Option{
		cmpopts.IgnoreFields(kong.Service{}, "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Route{}, "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Plugin{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.Upstream{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.Target{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.CACertificate{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.RBACEndpointPermission{}, "Role", "CreatedAt"),
		cmpopts.IgnoreFields(kong.RBACRole{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.Consumer{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.Vault{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Certificate{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.SNI{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.ConsumerGroup{}, "CreatedAt", "ID"),
		cmpopts.IgnoreFields(kong.ConsumerGroupPlugin{}, "CreatedAt", "ID"),
		cmpopts.IgnoreFields(kong.KeyAuth{}, "ID", "CreatedAt"),
		cmpopts.SortSlices(sortSlices),
		cmpopts.SortSlices(func(a, b *string) bool { return *a < *b }),
		cmpopts.EquateEmpty(),
	}
	opt = append(opt, ignoreFields...)

	if diff := cmp.Diff(kongState, &expectedState, opt...); diff != "" {
		t.Errorf(diff)
	}
}

func reset(t *testing.T, opts ...string) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"reset", "--force"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	if err := deckCmd.Execute(); err != nil {
		t.Fatalf(err.Error(), "failed to reset Kong's state")
	}
}

func readFile(filepath string) (string, error) {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func setup(_ *testing.T) func(t *testing.T) {
	// disable analytics for integration tests
	os.Setenv("DECK_ANALYTICS", "off")
	return func(t *testing.T) {
		reset(t)
	}
}

func sync(kongFile string, opts ...string) error {
	deckCmd := cmd.NewRootCmd()
	args := []string{"sync", "-s", kongFile}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	return deckCmd.ExecuteContext(context.Background())
}

func diff(kongFile string, opts ...string) (string, error) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"diff", "-s", kongFile}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)

	// overwrite default standard output
	r, w, _ := os.Pipe()
	color.Output = w

	// execute decK command
	cmdErr := deckCmd.ExecuteContext(context.Background())

	// read command output
	w.Close()
	out, _ := ioutil.ReadAll(r)

	return string(out), cmdErr
}

func dump(opts ...string) (string, error) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"dump"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)

	// capture command output to be used during tests
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmdErr := deckCmd.ExecuteContext(context.Background())

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	return string(out), cmdErr
}

func ping(opts ...string) error {
	deckCmd := cmd.NewRootCmd()
	args := []string{"ping"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	return deckCmd.ExecuteContext(context.Background())
}
