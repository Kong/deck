//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/kong/deck/utils"
	"github.com/stretchr/testify/assert"
)

var (
	expectedOutputMasked = `creating workspace test
creating service svc1  {
+  "connect_timeout": 60000
+  "host": "[masked]"
+  "id": "9ecf5708-f2f4-444e-a4c7-fcd3a57f9a6d"
+  "name": "svc1"
+  "protocol": "http"
+  "read_timeout": 60000
+  "tags": [
+    "[masked] is an external host. I like [masked]!",
+    "foo:foo",
+    "baz:[masked]",
+    "another:[masked]",
+    "bar:[masked]"
+  ]
+  "write_timeout": 60000
 }

creating plugin rate-limiting (global)  {
+  "config": {
+    "minute": 123
+  }
+  "id": "a1368a28-cb5c-4eee-86d8-03a6bdf94b5e"
+  "name": "rate-limiting"
 }

Summary:
  Created: 2
  Updated: 0
  Deleted: 0
`

	expectedOutputUnMasked = `creating workspace test
creating service svc1  {
+  "connect_timeout": 60000
+  "host": "mockbin.org"
+  "id": "9ecf5708-f2f4-444e-a4c7-fcd3a57f9a6d"
+  "name": "svc1"
+  "protocol": "http"
+  "read_timeout": 60000
+  "tags": [
+    "mockbin.org is an external host. I like mockbin.org!",
+    "foo:foo",
+    "baz:bazbaz",
+    "another:bazbaz",
+    "bar:barbar"
+  ]
+  "write_timeout": 60000
 }

creating plugin rate-limiting (global)  {
+  "config": {
+    "minute": 123
+  }
+  "id": "a1368a28-cb5c-4eee-86d8-03a6bdf94b5e"
+  "name": "rate-limiting"
 }

Summary:
  Created: 2
  Updated: 0
  Deleted: 0
`

	diffEnvVars = map[string]string{
		"DECK_SVC1_HOSTNAME": "mockbin.org",
		"DECK_BARR":          "barbar",
		"DECK_BAZZ":          "bazbaz",   // used more than once
		"DECK_FUB":           "fubfub",   // unused
		"DECK_FOO":           "foo_test", // unused, partial match
	}
)

// test scope:
//   - 1.x
//   - 2.x
func Test_Diff_Workspace_UnMasked_OlderThan3x(t *testing.T) {
	tests := []struct {
		name          string
		stateFile     string
		expectedState utils.KongRawState
		envVars       map[string]string
	}{
		{
			name:      "diff with not existent workspace doesn't error out",
			stateFile: "testdata/diff/001-not-existing-workspace/kong.yaml",
			envVars:   diffEnvVars,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				os.Setenv(k, v)
				defer func(k string) {
					os.Unsetenv(k)
				}(k)
			}
			runWhen(t, "kong", "<3.0.0")
			teardown := setup(t)
			defer teardown(t)

			out, err := diff(tc.stateFile, "--no-mask-deck-env-vars-value")
			assert.NoError(t, err)
			assert.Equal(t, out, expectedOutputUnMasked)
		})
	}
}
func Test_Diff_Workspace_Masked_OlderThan3x(t *testing.T) {
	tests := []struct {
		name          string
		stateFile     string
		expectedState utils.KongRawState
		envVars       map[string]string
	}{
		{
			name:      "diff with not existent workspace doesn't error out",
			stateFile: "testdata/diff/001-not-existing-workspace/kong.yaml",
			envVars:   diffEnvVars,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				os.Setenv(k, v)
				defer func(k string) {
					os.Unsetenv(k)
				}(k)
			}
			runWhen(t, "kong", "<3.0.0")
			teardown := setup(t)
			defer teardown(t)

			out, err := diff(tc.stateFile)
			assert.NoError(t, err)
			assert.Equal(t, out, expectedOutputMasked)
		})
	}
}

// test scope:
//   - 3.x
func Test_Diff_Workspace_UnMasked_NewerThan3x(t *testing.T) {
	tests := []struct {
		name          string
		stateFile     string
		expectedState utils.KongRawState
		envVars       map[string]string
	}{
		{
			name:      "diff with not existent workspace doesn't error out",
			stateFile: "testdata/diff/001-not-existing-workspace/kong3x.yaml",
			envVars:   diffEnvVars,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				os.Setenv(k, v)
				defer func(k string) {
					os.Unsetenv(k)
				}(k)
			}
			runWhen(t, "kong", ">=3.0.0")
			teardown := setup(t)
			defer teardown(t)

			out, err := diff(tc.stateFile, "--no-mask-deck-env-vars-value")
			assert.NoError(t, err)
			assert.Equal(t, out, expectedOutputUnMasked)
		})
	}
}
func Test_Diff_Workspace_Masked_NewerThan3x(t *testing.T) {
	tests := []struct {
		name          string
		stateFile     string
		expectedState utils.KongRawState
		envVars       map[string]string
	}{
		{
			name:      "diff with not existent workspace doesn't error out",
			stateFile: "testdata/diff/001-not-existing-workspace/kong3x.yaml",
			envVars:   diffEnvVars,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				os.Setenv(k, v)
				defer func(k string) {
					os.Unsetenv(k)
				}(k)
			}
			runWhen(t, "kong", ">=3.0.0")
			teardown := setup(t)
			defer teardown(t)

			out, err := diff(tc.stateFile)
			assert.NoError(t, err)
			assert.Equal(t, out, expectedOutputMasked)
		})
	}
}
