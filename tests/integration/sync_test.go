//go:build integration

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

var (
	// missing Enable
	svc1 = []*kong.Service{
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
	}

	// latest
	svc1_207 = []*kong.Service{
		{
			Name:           kong.String("svc1"),
			ConnectTimeout: kong.Int(60000),
			Host:           kong.String("mockbin.org"),
			Port:           kong.Int(80),
			Protocol:       kong.String("http"),
			ReadTimeout:    kong.Int(60000),
			Retries:        kong.Int(5),
			WriteTimeout:   kong.Int(60000),
			Enabled:        kong.Bool(true),
			Tags:           nil,
		},
	}

	// missing RequestBuffering, ResponseBuffering, Service, PathHandling
	route1_143 = []*kong.Route{
		{
			Name:                    kong.String("r1"),
			Paths:                   []*string{kong.String("/r1")},
			PreserveHost:            kong.Bool(false),
			Protocols:               []*string{kong.String("http"), kong.String("https")},
			RegexPriority:           kong.Int(0),
			StripPath:               kong.Bool(false),
			HTTPSRedirectStatusCode: kong.Int(301),
		},
	}

	// missing RequestBuffering, ResponseBuffering
	// PathHandling set to v1
	route1_151 = []*kong.Route{
		{
			Name:                    kong.String("r1"),
			Paths:                   []*string{kong.String("/r1")},
			PathHandling:            kong.String("v1"),
			PreserveHost:            kong.Bool(false),
			Protocols:               []*string{kong.String("http"), kong.String("https")},
			RegexPriority:           kong.Int(0),
			StripPath:               kong.Bool(false),
			HTTPSRedirectStatusCode: kong.Int(301),
			Service: &kong.Service{
				ID: kong.String("6d4e90fa-cb78-4607-8c4f-f12245ba8b59"),
			},
		},
	}

	// missing RequestBuffering, ResponseBuffering
	route1_205_214 = []*kong.Route{
		{
			Name:                    kong.String("r1"),
			Paths:                   []*string{kong.String("/r1")},
			PathHandling:            kong.String("v0"),
			PreserveHost:            kong.Bool(false),
			Protocols:               []*string{kong.String("http"), kong.String("https")},
			RegexPriority:           kong.Int(0),
			StripPath:               kong.Bool(false),
			HTTPSRedirectStatusCode: kong.Int(301),
			Service: &kong.Service{
				ID: kong.String("6d4e90fa-cb78-4607-8c4f-f12245ba8b59"),
			},
		},
	}

	// latest
	route1_20x = []*kong.Route{
		{
			Name:                    kong.String("r1"),
			Paths:                   []*string{kong.String("/r1")},
			PathHandling:            kong.String("v0"),
			PreserveHost:            kong.Bool(false),
			Protocols:               []*string{kong.String("http"), kong.String("https")},
			RegexPriority:           kong.Int(0),
			StripPath:               kong.Bool(false),
			HTTPSRedirectStatusCode: kong.Int(301),
			RequestBuffering:        kong.Bool(true),
			ResponseBuffering:       kong.Bool(true),
			Service: &kong.Service{
				ID: kong.String("8076db2-28b6-423b-ba39-a797193017f7"),
			},
		},
	}

	upstream = []*kong.Upstream{
		{
			Name:      kong.String("upstream1"),
			Algorithm: kong.String("round-robin"),
			Slots:     kong.Int(10000),
			Healthchecks: &kong.Healthcheck{
				Threshold: kong.Float64(0),
				Active: &kong.ActiveHealthcheck{
					Concurrency: kong.Int(10),
					Healthy: &kong.Healthy{
						HTTPStatuses: []int{200, 302},
						Interval:     kong.Int(0),
						Successes:    kong.Int(0),
					},
					HTTPPath:               kong.String("/"),
					Type:                   kong.String("http"),
					Timeout:                kong.Int(1),
					HTTPSVerifyCertificate: kong.Bool(true),
					Unhealthy: &kong.Unhealthy{
						HTTPFailures: kong.Int(0),
						TCPFailures:  kong.Int(0),
						Timeouts:     kong.Int(0),
						Interval:     kong.Int(0),
						HTTPStatuses: []int{429, 404, 500, 501, 502, 503, 504, 505},
					},
				},
				Passive: &kong.PassiveHealthcheck{
					Healthy: &kong.Healthy{
						HTTPStatuses: []int{
							200, 201, 202, 203, 204, 205,
							206, 207, 208, 226, 300, 301, 302, 303, 304, 305,
							306, 307, 308,
						},
						Successes: kong.Int(0),
					},
					Type: kong.String("http"),
					Unhealthy: &kong.Unhealthy{
						HTTPFailures: kong.Int(0),
						TCPFailures:  kong.Int(0),
						Timeouts:     kong.Int(0),
						HTTPStatuses: []int{429, 500, 503},
					},
				},
			},
			HashOn:           kong.String("none"),
			HashFallback:     kong.String("none"),
			HashOnCookiePath: kong.String("/"),
		},
	}

	target = []*kong.Target{
		{
			Target: kong.String("198.51.100.11:80"),
			Upstream: &kong.Upstream{
				ID: kong.String("a6f89ffc-1e53-4b01-9d3d-7a142bcd"),
			},
			Weight: kong.Int(100),
		},
	}

	targetZeroWeight = []*kong.Target{
		{
			Target: kong.String("198.51.100.11:80"),
			Upstream: &kong.Upstream{
				ID: kong.String("a6f89ffc-1e53-4b01-9d3d-7a142bcd"),
			},
			Weight: kong.Int(0),
		},
	}
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
	expectedState utils.KongRawState, ignoreFields []cmp.Option) {
	// Get entities from Kong
	ctx := context.Background()
	kongState, err := dump.Get(ctx, client, dump.Config{})
	if err != nil {
		t.Errorf(err.Error())
	}

	opt := []cmp.Option{
		cmpopts.IgnoreFields(kong.Service{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Route{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Plugin{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.Upstream{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.Target{}, "ID", "CreatedAt"),
		cmpopts.SortSlices(sortSlices),
		cmpopts.EquateEmpty(),
	}
	opt = append(opt, ignoreFields...)

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

func sync(kongFile string) {
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

// test scope:
//   - 1.4.3
func Test_Sync_ServicesRoutes_Till_1_4_3(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	if err != nil {
		t.Errorf(err.Error())
	}

	// ignore entities fields based on Kong version
	ignoreFields := []cmp.Option{
		cmpopts.IgnoreFields(kong.Route{}, "Service"),
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
				Services: svc1,
			},
		},
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/002-create-services-and-routes/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1,
				Routes:   route1_143,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, "<=1.4.3")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, ignoreFields)
		})
	}
}

// test scope:
//   - 1.5.1
//   - 1.5.0.11+enterprise
func Test_Sync_ServicesRoutes_Till_1_5_1(t *testing.T) {
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
				Services: svc1,
			},
		},
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/002-create-services-and-routes/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1,
				Routes:   route1_151,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, ">1.4.3 <=1.5.1")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.0.5
//   - 2.1.4
func Test_Sync_ServicesRoutes_From_2_0_5_To_2_1_4(t *testing.T) {
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
				Services: svc1,
			},
		},
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/002-create-services-and-routes/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1,
				Routes:   route1_205_214,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, ">=2.0.5 <=2.1.4")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.2.2
//   - 2.3.3
//   - 2.4.1
//   - 2.5.1
//   - 2.6.0
//   - 2.2.1.3+enterprise
//   - 2.3.3.4+enterprise
//   - 2.4.1.3+enterprise
//   - 2.5.1.2+enterprise
func Test_Sync_ServicesRoutes_From_2_2_1_to_2_6_0(t *testing.T) {
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
				Services: svc1,
			},
		},
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/002-create-services-and-routes/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1,
				Routes:   route1_20x,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, ">2.2.1 <=2.6.0")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.7.0
//   - 2.6.0.2+enterprise
//   - 2.7.0.0+enterprise
func Test_Sync_ServicesRoutes_From_2_6_9(t *testing.T) {
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
				Services: svc1_207,
			},
		},
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/002-create-services-and-routes/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1_207,
				Routes:   route1_20x,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, ">2.6.9")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 1.5.0.11+enterprise
func Test_Sync_BasicAuth_Plugin_Earlier_Than_1_5_1(t *testing.T) {
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
			kongFile: "testdata/sync/003-create-a-plugin/kong.yaml",
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
			kong.RunWhenKong(t, "<1.5.1 !1.4.3")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 1.5.1
func Test_Sync_BasicAuth_Plugin_1_5_1(t *testing.T) {
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
			kongFile: "testdata/sync/003-create-a-plugin/kong.yaml",
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
						RunOn: kong.String("first"),
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, "==1.5.1")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.0.5
//   - 2.1.4
//   - 2.2.2
//   - 2.3.3
//   - 2.4.1
//   - 2.5.1
//   - 2.6.0
//   - 2.7.0
//   - 2.1.4.6+enterprise
//   - 2.2.1.3+enterprise
//   - 2.3.3.4+enterprise
//   - 2.4.1.3+enterprise
//   - 2.5.1.2+enterprise
//   - 2.6.0.2+enterprise
//   - 2.7.0.0+enterprise
func Test_Sync_BasicAuth_Plugin_From_2_0_5(t *testing.T) {
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
			kongFile: "testdata/sync/003-create-a-plugin/kong.yaml",
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
			kong.RunWhenKong(t, ">=2.0.5")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 1.4.3
//   - 1.5.1
//   - 1.5.0.11+enterprise
func Test_Sync_Upstream_Target_Till_1_5_2(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	if err != nil {
		t.Errorf(err.Error())
	}

	// ignore entities fields based on Kong version
	ignoreFields := []cmp.Option{
		cmpopts.IgnoreFields(kong.Healthcheck{}, "Threshold"),
	}

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates an upstream and target",
			kongFile: "testdata/sync/004-create-upstream-and-target/kong.yaml",
			expectedState: utils.KongRawState{
				Upstreams: upstream,
				Targets:   target,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, "<=1.5.2")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, ignoreFields)
		})
	}
}

// test scope:
//   - 2.0.5
//   - 2.1.4
//   - 2.2.2
//   - 2.3.3
//   - 2.4.1
//   - 2.5.1
//   - 2.6.0
//   - 2.7.0
//   - 2.1.4.6+enterprise
//   - 2.2.1.3+enterprise
//   - 2.3.3.4+enterprise
//   - 2.4.1.3+enterprise
//   - 2.5.1.2+enterprise
//   - 2.6.0.2+enterprise
//   - 2.7.0.0+enterprise
func Test_Sync_Upstream_Target_From_2x(t *testing.T) {
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
			name:     "creates an upstream and target",
			kongFile: "testdata/sync/004-create-upstream-and-target/kong.yaml",
			expectedState: utils.KongRawState{
				Upstreams: upstream,
				Targets:   target,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, ">=2.1.0")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.4.1
//   - 2.5.1
//   - 2.6.0
//   - 2.7.0
//   - 2.4.1.3+enterprise
//   - 2.5.1.2+enterprise
//   - 2.6.0.2+enterprise
//   - 2.7.0.0+enterprise
func Test_Sync_Upstreams_Target_ZeroWeight(t *testing.T) {
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
			name:     "creates an upstream and target with weight equals to zero",
			kongFile: "testdata/sync/005-create-upstream-and-target-weight/kong.yaml",
			expectedState: utils.KongRawState{
				Upstreams: upstream,
				Targets:   targetZeroWeight,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, ">=2.4.1")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}
