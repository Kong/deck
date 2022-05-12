//go:build integration

package integration

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

var (
	// missing Enable
	svc1 = []*kong.Service{
		{
			ID:             kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
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
			ID:             kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
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
			ID:                      kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			Name:                    kong.String("r1"),
			Paths:                   []*string{kong.String("/r1")},
			PreserveHost:            kong.Bool(false),
			Protocols:               []*string{kong.String("http"), kong.String("https")},
			RegexPriority:           kong.Int(0),
			StripPath:               kong.Bool(true),
			HTTPSRedirectStatusCode: kong.Int(301),
		},
	}

	// missing RequestBuffering, ResponseBuffering
	// PathHandling set to v1
	route1_151 = []*kong.Route{
		{
			ID:                      kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			Name:                    kong.String("r1"),
			Paths:                   []*string{kong.String("/r1")},
			PathHandling:            kong.String("v1"),
			PreserveHost:            kong.Bool(false),
			Protocols:               []*string{kong.String("http"), kong.String("https")},
			RegexPriority:           kong.Int(0),
			StripPath:               kong.Bool(true),
			HTTPSRedirectStatusCode: kong.Int(301),
			Service: &kong.Service{
				ID: kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			},
		},
	}

	// missing RequestBuffering, ResponseBuffering
	route1_205_214 = []*kong.Route{
		{
			ID:                      kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			Name:                    kong.String("r1"),
			Paths:                   []*string{kong.String("/r1")},
			PathHandling:            kong.String("v0"),
			PreserveHost:            kong.Bool(false),
			Protocols:               []*string{kong.String("http"), kong.String("https")},
			RegexPriority:           kong.Int(0),
			StripPath:               kong.Bool(true),
			HTTPSRedirectStatusCode: kong.Int(301),
			Service: &kong.Service{
				ID: kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			},
		},
	}

	// latest
	route1_20x = []*kong.Route{
		{
			ID:                      kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			Name:                    kong.String("r1"),
			Paths:                   []*string{kong.String("/r1")},
			PathHandling:            kong.String("v0"),
			PreserveHost:            kong.Bool(false),
			Protocols:               []*string{kong.String("http"), kong.String("https")},
			RegexPriority:           kong.Int(0),
			StripPath:               kong.Bool(true),
			HTTPSRedirectStatusCode: kong.Int(301),
			RequestBuffering:        kong.Bool(true),
			ResponseBuffering:       kong.Bool(true),
			Service: &kong.Service{
				ID: kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			},
		},
	}

	// has run-on set to 'first'
	plugin_143_151 = []*kong.Plugin{
		{
			Name: kong.String("basic-auth"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"anonymous":        "58076db2-28b6-423b-ba39-a797193017f7",
				"hide_credentials": false,
			},
			RunOn: kong.String("first"),
		},
	}

	// latest
	plugin = []*kong.Plugin{
		{
			Name: kong.String("basic-auth"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"anonymous":        "58076db2-28b6-423b-ba39-a797193017f7",
				"hide_credentials": false,
			},
		},
	}

	plugin_on_entities = []*kong.Plugin{
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"per_consumer": false,
			},
			Service: &kong.Service{
				ID: kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			},
		},
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"per_consumer": false,
			},
			Route: &kong.Route{
				ID: kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			},
		},
		{
			Name: kong.String("rate-limiting"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"day":                 nil,
				"fault_tolerant":      true,
				"header_name":         nil,
				"hide_client_headers": false,
				"hour":                float64(10),
				"limit_by":            "consumer",
				"minute":              nil,
				"month":               nil,
				"path":                nil,
				"policy":              "cluster",
				"redis_username":      nil,
				"redis_database":      float64(0),
				"redis_host":          nil,
				"redis_password":      nil,
				"redis_port":          float64(6379),
				"redis_server_name":   nil,
				"redis_ssl":           false,
				"redis_ssl_verify":    false,
				"redis_timeout":       float64(2000),
				"second":              nil,
				"year":                nil,
			},
			Consumer: &kong.Consumer{
				ID: kong.String("d2965b9b-0608-4458-a9f8-0b93d88d03b8"),
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

	rateLimitingPlugin = []*kong.Plugin{
		{
			Name: kong.String("rate-limiting"),
			Config: kong.Configuration{
				"day":                 nil,
				"fault_tolerant":      true,
				"header_name":         nil,
				"hide_client_headers": false,
				"hour":                nil,
				"limit_by":            "consumer",
				"minute":              float64(123),
				"month":               nil,
				"path":                nil,
				"policy":              "cluster",
				"redis_database":      float64(0),
				"redis_host":          nil,
				"redis_password":      nil,
				"redis_port":          float64(6379),
				"redis_server_name":   nil,
				"redis_ssl":           false,
				"redis_ssl_verify":    false,
				"redis_timeout":       float64(2000),
				"second":              nil,
				"year":                nil,
			},
			Enabled: kong.Bool(true),
			RunOn:   nil,
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Tags: nil,
		},
	}

	consumer = []*kong.Consumer{
		{
			Username: kong.String("yolo"),
			ID:       kong.String("d2965b9b-0608-4458-a9f8-0b93d88d03b8"),
		},
	}
)

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
//   - 1.4.3
func Test_Sync_BasicAuth_Plugin_1_4_3(t *testing.T) {
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
				Plugins: plugin_143_151,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, "==1.4.3")
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
				Plugins: plugin,
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
				Plugins: plugin_143_151,
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
				Plugins: plugin,
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

func Test_Sync_RateLimitingPlugin(t *testing.T) {
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
			name:     "fill defaults",
			kongFile: "testdata/sync/006-fill-defaults-rate-limiting/kong.yaml",
			expectedState: utils.KongRawState{
				Plugins: rateLimitingPlugin,
			},
		},
		{
			name:     "fill defaults with dedup",
			kongFile: "testdata/sync/007-fill-defaults-rate-limiting-dedup/kong.yaml",
			expectedState: utils.KongRawState{
				Plugins: rateLimitingPlugin,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, "==2.7.0")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 1.5.0.11+enterprise
func Test_Sync_FillDefaults_Earlier_Than_1_5_1(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	if err != nil {
		t.Errorf(err.Error())
	}

	// ignore entities fields based on Kong version
	ignoreFields := []cmp.Option{
		cmpopts.IgnoreFields(kong.Route{}, "Service"),
		cmpopts.IgnoreFields(kong.Healthcheck{}, "Threshold"),
	}

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates a service",
			kongFile: "testdata/sync/008-create-simple-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1,
				Routes:    route1_151,
				Plugins:   plugin,
				Targets:   target,
				Upstreams: upstream,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, "<1.5.1 !1.4.3")
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
func Test_Sync_FillDefaults_From_2_0_5_To_2_1_4(t *testing.T) {
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
			name:     "create services and routes",
			kongFile: "testdata/sync/008-create-simple-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1,
				Routes:    route1_205_214,
				Upstreams: upstream,
				Targets:   target,
				Plugins:   plugin,
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
func Test_Sync_FillDefaults_From_2_2_1_to_2_6_0(t *testing.T) {
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
			name:     "create services and routes",
			kongFile: "testdata/sync/008-create-simple-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1,
				Routes:    route1_20x,
				Upstreams: upstream,
				Targets:   target,
				Plugins:   plugin,
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
func Test_Sync_FillDefaults_From_2_6_9(t *testing.T) {
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
			name:     "creates entities with minimum configuration",
			kongFile: "testdata/sync/008-create-simple-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1_207,
				Routes:    route1_20x,
				Plugins:   plugin,
				Targets:   target,
				Upstreams: upstream,
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

func Test_Sync_SkipCACert(t *testing.T) {
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
			name:     "syncing with --skip-ca-certificates should ignore CA certs",
			kongFile: "testdata/sync/009-skip-ca-cert/kong.yaml",
			expectedState: utils.KongRawState{
				Services:       svc1_207,
				CACertificates: []*kong.CACertificate{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// ca_certificates first appeared in 1.3, but we limit to 2.7+
			// here because the schema changed and the entities aren't the same
			// across all versions, even though the skip functionality works the same.
			kong.RunWhenKong(t, ">=2.7.0")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile, "--skip-ca-certificates")
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}

func Test_Sync_RBAC(t *testing.T) {
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
			name:     "rbac",
			kongFile: "testdata/sync/xxx-rbac-endpoint-permissions/kong.yaml",
			expectedState: utils.KongRawState{
				RBACRoles: []*kong.RBACRole{
					{
						Name:    kong.String("workspace-portal-admin"),
						Comment: kong.String("Full access to Dev Portal related endpoints in the workspace"),
					},
				},
				RBACEndpointPermissions: []*kong.RBACEndpointPermission{
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/developers"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/developers/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/files"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/files/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/kong"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*/*/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*/*/*/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*/*/*/*/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/workspaces/default"),
						Actions:   []*string{kong.String("read"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenEnterprise(t, ">=2.7.0", kong.RequiredFeatures{})
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile, "--rbac-resources-only")
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}

func Test_Sync_Create_Route_With_Service_Name_Reference(t *testing.T) {
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
			name:     "create a route with a service name reference",
			kongFile: "testdata/sync/010-create-route-with-service-name-reference/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1_207,
				Routes:   route1_20x,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, ">=2.7.0")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}

func Test_Sync_PluginsOnEntities(t *testing.T) {
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
			name:     "create plugins on services, routes and consumers",
			kongFile: "testdata/sync/xxx-plugins-on-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1_207,
				Routes:    route1_20x,
				Plugins:   plugin_on_entities,
				Consumers: consumer,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, ">=2.8.0")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}
