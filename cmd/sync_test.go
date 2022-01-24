package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

func getKongAddress() string {
	if os.Getenv("KONG_ADDRESS") != "" {
		return os.Getenv("KONG_ADDRESS")
	}
	return "http://localhost:8001"
}

// cleanUpEnv removes all existing entities from Kong.
func cleanUpEnv(client *kong.Client) error {
	ctx := context.Background()
	currentState, err := fetchCurrentState(ctx, client, dumpConfig)
	if err != nil {
		return err
	}
	targetState, err := state.NewKongState()
	if err != nil {
		return err
	}
	_, err = performDiff(ctx, currentState, targetState, false, 10, 0, client)
	return err
}

func normalizeOutput(content string) string {
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	sort.Strings(lines)
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.Join(lines, "\n")
}

func loadExpectedOutput(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func testDeckOutput(t *testing.T, outputPath string, got string) {
	expected, err := loadExpectedOutput(outputPath)
	if err != nil {
		t.Errorf(err.Error())
	}
	expectedNormalized := normalizeOutput(expected)
	obtainedNormalized := normalizeOutput(got)
	if !reflect.DeepEqual(obtainedNormalized, expectedNormalized) {
		t.Errorf(cmp.Diff(obtainedNormalized, expectedNormalized))
	}
}

func setupTest() (*kong.Client, string, error) {
	var client *kong.Client
	var kongVersion string

	config := utils.KongClientConfig{
		Address: getKongAddress(),
	}
	client, err := utils.GetKongClient(config)
	if err != nil {
		return client, kongVersion, err
	}

	if err := cleanUpEnv(client); err != nil {
		return client, kongVersion, err
	}

	kongVersion, err = fetchKongVersion(context.Background(), config)
	if err != nil {
		return client, kongVersion, err
	}
	return client, kongVersion, nil
}

func Test_Sync_output(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create_service",
		},
		{
			name: "modify_service",
		},
		{
			name: "create_service_no_change",
		},
		{
			name: "create_route",
		},
		{
			name: "modify_route",
		},
		{
			name: "create_route_no_change",
		},
		{
			name: "create_plugin",
		},
		{
			name: "modify_plugin",
		},
		{
			name: "create_plugin_no_change",
		},
	}
	_, kongVersion, err := setupTest()
	if err != nil {
		t.Errorf(err.Error())
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tcPath := fmt.Sprintf("testdata/sync/%s/%s", kongVersion, tc.name)
			kongFile := fmt.Sprintf("%s/kong.yaml", tcPath)
			cmd := exec.Command(
				"../deck",
				"--kong-addr",
				getKongAddress(),
				"sync",
				"-s",
				kongFile,
			) // #nosec G204

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Errorf(stderr.String())
			}

			// Check deck output looks as expected.
			testDeckOutput(t, fmt.Sprintf("%s/output", tcPath), stdout.String())
		})
	}
}
