//go:build integration

package e2e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var DECK = filepath.Join("../../deck")

func getKongAddress() string {
	address := os.Getenv("KONG_ADDRESS")
	if address != "" {
		return address
	}
	return "http://localhost:8001"
}

// cleanUpEnv removes all existing entities from Kong.
func cleanUpEnv() error {
	cmd := exec.Command(
		DECK,
		"--kong-addr",
		getKongAddress(),
		"reset",
		"--force",
	) // #nosec G204

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(stderr.String())
	}
	return nil
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

func Test_Sync_output(t *testing.T) {
	tests := []struct {
		name               string
		kongFile           string
		expectedOutputFile string
	}{
		{
			name:               "create_service",
			kongFile:           "testdata/sync/create_service/kong.yaml",
			expectedOutputFile: "testdata/sync/create_service/output",
		},
	}

	if err := cleanUpEnv(); err != nil {
		t.Errorf(err.Error())
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(
				DECK,
				"--kong-addr",
				getKongAddress(),
				"sync",
				"-s",
				tc.kongFile,
			) // #nosec G204

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Errorf(stderr.String())
			}

			// Check deck output looks as expected.
			testDeckOutput(t, tc.expectedOutputFile, stdout.String())
		})
	}
}
