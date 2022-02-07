//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kong/deck/cmd"
	"github.com/kong/deck/dump"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var deckCmd = cmd.RootCmdOnlyForDocsAndTest

var syncCmd = func() *cobra.Command {
	for _, command := range deckCmd.Commands() {
		if command.Use == "sync" {
			return command
		}
	}
	return nil
}

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

func testKongState(t *testing.T, client *kong.Client, expectedState utils.KongRawState) {
	// Get entities from Kong
	ctx := context.Background()
	kongState, err := dump.Get(ctx, client, dump.Config{})
	if err != nil {
		t.Errorf(err.Error())
	}

	opt := []cmp.Option{
		cmpopts.IgnoreFields(kong.Service{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Route{}, "ID", "CreatedAt", "UpdatedAt",
			"Service", "RequestBuffering", "ResponseBuffering", "PathHandling"),
		cmpopts.IgnoreFields(kong.Plugin{}, "ID", "CreatedAt"),
		cmpopts.SortSlices(sortSlices),
		cmpopts.EquateEmpty(),
	}

	if diff := cmp.Diff(kongState, &expectedState, opt...); diff != "" {
		t.Errorf(diff)
	}
}

func reset(t *testing.T) {
	deckCmd.SetArgs([]string{"reset", "--force"})
	if err := deckCmd.Execute(); err != nil {
		t.Fatalf(err.Error(), "failed to reset Kong's state")
	}
}

func setup(t *testing.T) func(t *testing.T) {
	return func(t *testing.T) {
		reset(t)
	}
}

func sync(t *testing.T, kongFile string) {
	// set the --state flag directly due to slice
	// flags value are persisted across test cases, and not
	// overwritable otherwise.
	stateFlag := syncCmd().Flags().Lookup("state")
	if val, ok := stateFlag.Value.(pflag.SliceValue); ok {
		_ = val.Replace([]string{kongFile})
	}

	deckCmd.SetArgs([]string{"sync"})
	deckCmd.Execute()
}

func Test_Sync_ServicesRoutes(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	if err != nil {
		t.Errorf(err.Error())
	}

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates a service",
			kongFile: "testdata/sync/001-create-a-service/kong.yaml",
			expectedState: utils.KongRawState{
				Services: []*kong.Service{
					{
						Name:           kong.String("svc1"),
						ConnectTimeout: kong.Int(60000),
						Host:           kong.String("mockbin.org"),
						Port:           kong.Int(80),
						Protocol:       kong.String("http"),
						ReadTimeout:    kong.Int(60000),
						Retries:        kong.Int(5),
						WriteTimeout:   kong.Int(60000),
						Tags:           nil,
					},
				},
			},
		},
		{
			name:     "create multiple services",
			kongFile: "testdata/sync/002-create-multiple-services/kong.yaml",
			expectedState: utils.KongRawState{
				Services: []*kong.Service{
					{
						Name:           kong.String("svc1"),
						ConnectTimeout: kong.Int(60000),
						Host:           kong.String("mockbin.org"),
						Port:           kong.Int(80),
						Protocol:       kong.String("http"),
						ReadTimeout:    kong.Int(60000),
						Retries:        kong.Int(5),
						WriteTimeout:   kong.Int(60000),
						Tags:           nil,
					},
					{
						Name:           kong.String("svc2"),
						ConnectTimeout: kong.Int(60000),
						Host:           kong.String("mockbin-v2.org"),
						Port:           kong.Int(8080),
						Protocol:       kong.String("https"),
						ReadTimeout:    kong.Int(60000),
						Retries:        kong.Int(5),
						WriteTimeout:   kong.Int(60000),
						Tags:           nil,
					},
				},
			},
		},
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/003-create-services-and-routes/kong.yaml",
			expectedState: utils.KongRawState{
				Services: []*kong.Service{
					{
						Name:           kong.String("svc1"),
						ConnectTimeout: kong.Int(60000),
						Host:           kong.String("mockbin.org"),
						Port:           kong.Int(80),
						Protocol:       kong.String("http"),
						ReadTimeout:    kong.Int(60000),
						Retries:        kong.Int(5),
						WriteTimeout:   kong.Int(60000),
						Tags:           nil,
					},
					{
						Name:           kong.String("svc2"),
						ConnectTimeout: kong.Int(60000),
						Host:           kong.String("mockbin-v2.org"),
						Port:           kong.Int(8080),
						Protocol:       kong.String("https"),
						ReadTimeout:    kong.Int(60000),
						Retries:        kong.Int(5),
						WriteTimeout:   kong.Int(60000),
						Tags:           nil,
					},
				},
				Routes: []*kong.Route{
					{
						Name:                    kong.String("r1"),
						Paths:                   []*string{kong.String("/r1")},
						PathHandling:            kong.String("v0"),
						PreserveHost:            kong.Bool(false),
						Protocols:               []*string{kong.String("http"), kong.String("https")},
						RegexPriority:           kong.Int(0),
						StripPath:               kong.Bool(false),
						HTTPSRedirectStatusCode: kong.Int(301),
					},
					{
						Name:                    kong.String("r2"),
						Paths:                   []*string{kong.String("/r2")},
						PathHandling:            kong.String("v0"),
						PreserveHost:            kong.Bool(false),
						Protocols:               []*string{kong.String("http"), kong.String("https")},
						RegexPriority:           kong.Int(0),
						StripPath:               kong.Bool(false),
						HTTPSRedirectStatusCode: kong.Int(301),
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown(t)

			sync(t, tc.kongFile)
			testKongState(t, client, tc.expectedState)
		})
	}
}

// runWhenKong skips the current test if the version of Kong doesn't
// fall in the semverRange.
// This helper function can be used in tests to write version specific
// tests for Kong.
func runWhenKong(t *testing.T, client *kong.Client, semverRange string) {
	// get kong version
	ctx := context.Background()
	info, err := client.Root(ctx)
	if err != nil {
		t.Error(err)
	}
	kongVersion := kong.VersionFromInfo(info)
	currentVersion, err := kong.ParseSemanticVersion(kongVersion)
	if err != nil {
		t.Error(err)
	}
	r, err := semver.ParseRange(semverRange)
	if err != nil {
		t.Error(err)
	}
	if !r(currentVersion) {
		t.Skip()
	}
}

func Test_Sync_Plugins(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	if err != nil {
		t.Errorf(err.Error())
	}

	tests := []struct {
		name            string
		kongFile        string
		initialKongFile string
		expectedState   utils.KongRawState
	}{
		{
			name:     "create a plugin",
			kongFile: "testdata/sync/004-create-a-plugin/kong.yaml",
			expectedState: utils.KongRawState{
				Plugins: []*kong.Plugin{
					{
						Name:      kong.String("basic-auth"),
						Protocols: []*string{kong.String("http"), kong.String("https")},
						Enabled:   kong.Bool(true),
						Config: kong.Configuration{
							"anonymous":        "58076db2-28b6-423b-ba39-a797193017f7",
							"hide_credentials": false,
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhenKong(t, client, ">=2.0.5")
			teardown := setup(t)
			defer teardown(t)

			sync(t, tc.kongFile)
			testKongState(t, client, tc.expectedState)
		})
	}
}
