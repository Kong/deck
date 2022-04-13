//nolint:deadcode
package integration

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kong/deck/cmd"
	"github.com/kong/deck/dump"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

func getKongAddress() string {
	address := os.Getenv("DECK_KONG_ADDR")
	if address != "" {
		return address
	}
	return "http://localhost:8001"
}

func getTestClient() (*kong.Client, error) {
	return utils.GetKongClient(utils.KongClientConfig{
		Address: getKongAddress(),
	})
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
	case *kong.Plugin:
		yEntity := y.(*kong.Plugin)
		xName = *xEntity.Name
		yName = *yEntity.Name
	}
	return xName < yName
}

func testKongState(t *testing.T, client *kong.Client,
	expectedState utils.KongRawState, ignoreFields []cmp.Option,
) {
	// Get entities from Kong
	ctx := context.Background()
	dumpConfig := dump.Config{}
	if expectedState.RBACEndpointPermissions != nil {
		dumpConfig.RBACResourcesOnly = true
	}
	kongState, err := dump.Get(ctx, client, dumpConfig)
	if err != nil {
		t.Errorf(err.Error())
	}

	opt := []cmp.Option{
		cmpopts.IgnoreFields(kong.Service{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Route{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Plugin{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.Upstream{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.Target{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.CACertificate{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.RBACEndpointPermission{}, "Role", "CreatedAt"),
		cmpopts.IgnoreFields(kong.RBACRole{}, "ID", "CreatedAt"),
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

func setup(t *testing.T) func(t *testing.T) {
	// disable analytics for integration tests
	os.Setenv("DECK_ANALYTICS", "off")
	return func(t *testing.T) {
		reset(t)
	}
}

func sync(kongFile string, opts ...string) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"sync", "-s", kongFile}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	deckCmd.ExecuteContext(context.Background()) //nolint:errcheck
}

func diff(kongFile string, opts ...string) error {
	deckCmd := cmd.NewRootCmd()
	args := []string{"diff", "-s", kongFile}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	return deckCmd.ExecuteContext(context.Background())
}
