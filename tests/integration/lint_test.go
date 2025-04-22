//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kong/deck/lint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func Test_LintPlain(t *testing.T) {
	tests := []struct {
		name        string
		stateFile   string
		rulesetFile string
	}{
		{
			name:        "lint plain",
			stateFile:   "testdata/lint/001-simple-lint/kong.yaml",
			rulesetFile: "testdata/lint/001-simple-lint/ruleset.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := file_lint(
				"-s", tc.stateFile,
				tc.rulesetFile,
			)
			require.Error(t, err)

			assert.Contains(t, output, "Linting Violations: 2")
			assert.Contains(t, output, "Failures: 1")
			assert.Contains(t, output, "[warn][5:13] Must use HTTPS protocol: `http` does not match the expression `^https`")
			assert.Contains(t, output, "[error][1:18] Check the version is correct: `3.0` does not match the expression `^1.1$`")
		})
	}
}

type lintErrors struct {
	TotalCount int
	FailCount  int
	Results    []lint.Result
}

func Test_LintStructured(t *testing.T) {
	tests := []struct {
		name                string
		stateFile           string
		rulesetFile         string
		expectedFile        string
		format              string
		displayOnlyFailures bool
		failSeverity        string
	}{
		{
			name:         "lint yaml",
			stateFile:    "testdata/lint/001-simple-lint/kong.yaml",
			rulesetFile:  "testdata/lint/001-simple-lint/ruleset.yaml",
			expectedFile: "testdata/lint/001-simple-lint/expected.yaml",
			format:       "yaml",
		},
		{
			name:                "lint yaml with modified severity",
			stateFile:           "testdata/lint/001-simple-lint/kong.yaml",
			rulesetFile:         "testdata/lint/001-simple-lint/ruleset.yaml",
			expectedFile:        "testdata/lint/001-simple-lint/expected-fail-severity-error.yaml",
			format:              "yaml",
			displayOnlyFailures: true,
			failSeverity:        "error",
		},
		{
			name:         "lint json",
			stateFile:    "testdata/lint/001-simple-lint/kong.yaml",
			rulesetFile:  "testdata/lint/001-simple-lint/ruleset.yaml",
			expectedFile: "testdata/lint/001-simple-lint/expected.json",
			format:       "json",
		},
		{
			name:                "lint json with modified severity",
			stateFile:           "testdata/lint/001-simple-lint/kong.yaml",
			rulesetFile:         "testdata/lint/001-simple-lint/ruleset.yaml",
			expectedFile:        "testdata/lint/001-simple-lint/expected-fail-severity-error.json",
			format:              "json",
			displayOnlyFailures: true,
			failSeverity:        "error",
		},
		{
			name:         "lint OAS with recommended ruleset",
			stateFile:    "testdata/lint/002-extends/oas.yaml",
			rulesetFile:  "testdata/lint/002-extends/ruleset-recommended-only.yaml",
			expectedFile: "testdata/lint/002-extends/expected-recommended-only.yaml",
			format:       "yaml",
			failSeverity: "info",
		},
		{
			name:         "lint decK with recommended ruleset OFF",
			stateFile:    "testdata/lint/002-extends/kong.yaml",
			rulesetFile:  "testdata/lint/002-extends/ruleset-recommended-off-kong.yaml",
			expectedFile: "testdata/lint/002-extends/expected-recommended-off-kong.yaml",
			format:       "yaml",
			failSeverity: "error",
		},
		{
			name:         "lint OAS with recommended ruleset AND custom rules",
			stateFile:    "testdata/lint/002-extends/oas.yaml",
			rulesetFile:  "testdata/lint/002-extends/ruleset-recommended-plus-custom.yaml",
			expectedFile: "testdata/lint/002-extends/expected-recommended-plus-custom.yaml",
			format:       "yaml",
			failSeverity: "error",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lintOpts := []string{
				"-s", tc.stateFile,
				"--format", tc.format,
				tc.rulesetFile,
			}
			if tc.displayOnlyFailures {
				lintOpts = append(lintOpts, "--display-only-failures")
			}
			if tc.failSeverity != "" {
				lintOpts = append(lintOpts, "--fail-severity", tc.failSeverity)
			}
			output, err := file_lint(lintOpts...)
			require.Error(t, err)

			var expectedErrors, outputErrors lintErrors
			// get expected errors from file
			content, err := os.ReadFile(tc.expectedFile)
			require.NoError(t, err)

			if tc.format == "yaml" {
				err = yaml.Unmarshal(content, &expectedErrors)
				require.NoError(t, err)

				// parse result errors from lint command
				err = yaml.Unmarshal([]byte(output), &outputErrors)
				require.NoError(t, err)
			} else {
				err = json.Unmarshal(content, &expectedErrors)
				require.NoError(t, err)

				// parse result errors from lint command
				err = json.Unmarshal([]byte(output), &outputErrors)
				require.NoError(t, err)
			}

			cmpOpts := []cmp.Option{
				cmpopts.SortSlices(func(a, b lint.Result) bool { return a.Line < b.Line }),
				cmpopts.EquateEmpty(),
			}
			if diff := cmp.Diff(outputErrors, expectedErrors, cmpOpts...); diff != "" {
				t.Errorf("got unexpected diff\n:%s", diff)
			}
		})
	}
}
