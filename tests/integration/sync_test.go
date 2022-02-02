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
)

var deckCmd = cmd.RootCmdOnlyForDocsAndTest

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

func testKongState(
	t *testing.T,
	client *kong.Client,
	expectedServices []*kong.Service,
	expectedRoutes []*kong.Route,
	expectedPlugins []*kong.Plugin,
) {
	ctx := context.Background()
	// Get entities from Kong
	services, err := dump.GetAllServices(ctx, client, []string{})
	if err != nil {
		t.Errorf(err.Error())
	}
	routes, err := dump.GetAllRoutes(ctx, client, []string{})
	if err != nil {
		t.Errorf(err.Error())
	}
	plugins, err := dump.GetAllPlugins(ctx, client, []string{})
	if err != nil {
		t.Errorf(err.Error())
	}
	opt := []cmp.Option{
		cmpopts.IgnoreFields(kong.Service{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Route{}, "ID", "CreatedAt", "UpdatedAt",
			"Service", "RequestBuffering", "ResponseBuffering", "PathHandling"),
		cmpopts.IgnoreFields(kong.Plugin{}, "ID", "CreatedAt"),
		cmpopts.SortSlices(sortSlices),
	}
	if diff := cmp.Diff(services, expectedServices, opt...); diff != "" {
		t.Errorf(diff)
	}
	if diff := cmp.Diff(routes, expectedRoutes, opt...); diff != "" {
		t.Errorf(diff)
	}
	if diff := cmp.Diff(plugins, expectedPlugins, opt...); diff != "" {
		t.Errorf(diff)
	}
}

func reset(t *testing.T) {
	deckCmd.SetArgs([]string{"reset", "--force"})
	if err := deckCmd.Execute(); err != nil {
		t.Errorf(err.Error())
	}
}

func setup(t *testing.T, kongFile string) func(t *testing.T) {
	if kongFile != "" {
		sync(t, kongFile)
	}
	return func(t *testing.T) {
		reset(t)
	}
}

func sync(t *testing.T, kongFile string) {
	// directly set flag value due to cobra keeping state across test runs
	cmd.SyncCmdKongStateFile = []string{kongFile}
	deckCmd.SetArgs([]string{"sync"})
	err := deckCmd.Execute()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test_Sync(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	if err != nil {
		t.Errorf(err.Error())
	}

	tests := []struct {
		name             string
		kongFile         string
		initialKongFile  string
		expectedServices []*kong.Service
		expectedRoutes   []*kong.Route
		expectedPlugins  []*kong.Plugin
	}{
		{
			name:     "creates a service",
			kongFile: "testdata/sync/create_service/kong.yaml",
			expectedServices: []*kong.Service{
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
		{
			name:            "service already exists",
			initialKongFile: "testdata/sync/create_service/kong.yaml",
			kongFile:        "testdata/sync/create_service/kong.yaml",
			expectedServices: []*kong.Service{
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
		{
			name:            "create new service",
			initialKongFile: "testdata/sync/create_new_service/base.yaml",
			kongFile:        "testdata/sync/create_new_service/kong.yaml",
			expectedServices: []*kong.Service{
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
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/create_services_and_routes/kong.yaml",
			expectedServices: []*kong.Service{
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
			expectedRoutes: []*kong.Route{
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
		// {
		// 	name:     "create services routes and plugins",
		// 	kongFile: "testdata/sync/create_services_routes_and_plugins/kong.yaml",
		// 	expectedServices: []*kong.Service{
		// 		{
		// 			Name:           kong.String("svc1"),
		// 			ConnectTimeout: kong.Int(60000),
		// 			Host:           kong.String("mockbin.org"),
		// 			Port:           kong.Int(80),
		// 			Protocol:       kong.String("http"),
		// 			ReadTimeout:    kong.Int(60000),
		// 			Retries:        kong.Int(5),
		// 			WriteTimeout:   kong.Int(60000),
		// 			Tags:           nil,
		// 		},
		// 		{
		// 			Name:           kong.String("svc2"),
		// 			ConnectTimeout: kong.Int(60000),
		// 			Host:           kong.String("mockbin-v2.org"),
		// 			Port:           kong.Int(8080),
		// 			Protocol:       kong.String("https"),
		// 			ReadTimeout:    kong.Int(60000),
		// 			Retries:        kong.Int(5),
		// 			WriteTimeout:   kong.Int(60000),
		// 			Tags:           nil,
		// 		},
		// 	},
		// 	expectedRoutes: []*kong.Route{
		// 		{
		// 			Name:                    kong.String("r1"),
		// 			Paths:                   []*string{kong.String("/r1")},
		// 			PathHandling:            kong.String("v0"),
		// 			PreserveHost:            kong.Bool(false),
		// 			Protocols:               []*string{kong.String("http"), kong.String("https")},
		// 			RegexPriority:           kong.Int(0),
		// 			StripPath:               kong.Bool(false),
		// 			HTTPSRedirectStatusCode: kong.Int(301),
		// 			RequestBuffering:        kong.Bool(true),
		// 			ResponseBuffering:       kong.Bool(true),
		// 		},
		// 		{
		// 			Name:                    kong.String("r2"),
		// 			Paths:                   []*string{kong.String("/r2")},
		// 			PathHandling:            kong.String("v0"),
		// 			PreserveHost:            kong.Bool(false),
		// 			Protocols:               []*string{kong.String("http"), kong.String("https")},
		// 			RegexPriority:           kong.Int(0),
		// 			StripPath:               kong.Bool(false),
		// 			HTTPSRedirectStatusCode: kong.Int(301),
		// 			RequestBuffering:        kong.Bool(true),
		// 			ResponseBuffering:       kong.Bool(true),
		// 		},
		// 	},
		// 	expectedPlugins: []*kong.Plugin{
		// 		{
		// 			Name:      kong.String("prometheus"),
		// 			Protocols: []*string{kong.String("http"), kong.String("https")},
		// 			Enabled:   kong.Bool(true),
		// 		},
		// 	},
		// },
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardown := setup(t, tc.initialKongFile)
			defer teardown(t)

			sync(t, tc.kongFile)
			testKongState(
				t, client,
				tc.expectedServices,
				tc.expectedRoutes,
				tc.expectedPlugins,
			)
		})
	}
}
