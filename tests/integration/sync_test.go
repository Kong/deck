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

// cleanUpEnv removes all existing entities from Kong.
func cleanUpEnv(t *testing.T) error {
	deckCmd.SetArgs([]string{"reset", "--force"})
	if err := deckCmd.Execute(); err != nil {
		t.Errorf(err.Error())
	}
	return nil
}

func getTestClient() (*kong.Client, error) {
	return utils.GetKongClient(utils.KongClientConfig{
		Address: getKongAddress(),
	})
}

func testKongState(
	t *testing.T,
	client *kong.Client,
	expectedServices []*kong.Service,
) {
	ctx := context.Background()
	// Get entities from Kong
	services, err := dump.GetAllServices(ctx, client, []string{})
	if err != nil {
		t.Errorf(err.Error())
	}
	opt := []cmp.Option{
		cmpopts.IgnoreFields(kong.Service{}, "ID", "CreatedAt", "UpdatedAt"),
	}
	if diff := cmp.Diff(services, expectedServices, opt...); diff != "" {
		t.Errorf(diff)
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
		expectedServices []*kong.Service
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := cleanUpEnv(t); err != nil {
				t.Errorf(err.Error())
			}

			deckCmd.SetArgs([]string{"sync", "-s", tc.kongFile})
			if err := deckCmd.Execute(); err != nil {
				t.Errorf(err.Error())
			}
			testKongState(t, client, tc.expectedServices)
		})
	}
}
