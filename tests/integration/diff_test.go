//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	emptyOutput = `Summary:
  Created: 0
  Updated: 0
  Deleted: 0
`

	expectedOutputMasked = `updating service svc1  {
   "connect_timeout": 60000,
   "enabled": true,
   "host": "[masked]",
   "id": "9ecf5708-f2f4-444e-a4c7-fcd3a57f9a6d",
   "name": "svc1",
   "port": 80,
   "protocol": "http",
   "read_timeout": 60000,
   "retries": 5,
   "write_timeout": 60000
+  "tags": [
+    "[masked] is an external host. I like [masked]!",
+    "foo:foo",
+    "baz:[masked]",
+    "another:[masked]",
+    "bar:[masked]"
+  ]
 }

creating plugin rate-limiting (global)
Summary:
  Created: 1
  Updated: 1
  Deleted: 0
`

	expectedOutputUnMasked = `updating service svc1  {
   "connect_timeout": 60000,
   "enabled": true,
   "host": "mockbin.org",
   "id": "9ecf5708-f2f4-444e-a4c7-fcd3a57f9a6d",
   "name": "svc1",
   "port": 80,
   "protocol": "http",
   "read_timeout": 60000,
   "retries": 5,
   "write_timeout": 60000
+  "tags": [
+    "test"
+  ]
 }

creating plugin rate-limiting (global)
Summary:
  Created: 1
  Updated: 1
  Deleted: 0
`

	diffEnvVars = map[string]string{
		"DECK_SVC1_HOSTNAME": "mockbin.org",
		"DECK_BARR":          "barbar",
		"DECK_BAZZ":          "bazbaz",   // used more than once
		"DECK_FUB":           "fubfub",   // unused
		"DECK_FOO":           "foo_test", // unused, partial match
	}

	expectedOutputUnMaskedJSON = `{
	"changes": {
		"creating": [
			{
				"name": "rate-limiting (global)",
				"kind": "plugin",
				"body": {
					"new": {
						"id": "a1368a28-cb5c-4eee-86d8-03a6bdf94b5e",
						"name": "rate-limiting",
						"config": {
							"day": null,
							"error_code": 429,
							"error_message": "API rate limit exceeded",
							"fault_tolerant": true,
							"header_name": null,
							"hide_client_headers": false,
							"hour": null,
							"limit_by": "consumer",
							"minute": 123,
							"month": null,
							"path": null,
							"policy": "local",
							"redis_database": 0,
							"redis_host": null,
							"redis_password": null,
							"redis_port": 6379,
							"redis_server_name": null,
							"redis_ssl": false,
							"redis_ssl_verify": false,
							"redis_timeout": 2000,
							"redis_username": null,
							"second": null,
							"year": null
						},
						"enabled": true,
						"protocols": [
							"grpc",
							"grpcs",
							"http",
							"https"
						]
					},
					"old": null
				}
			}
		],
		"updating": [
			{
				"name": "svc1",
				"kind": "service",
				"body": {
					"new": {
						"connect_timeout": 60000,
						"enabled": true,
						"host": "mockbin.org",
						"id": "9ecf5708-f2f4-444e-a4c7-fcd3a57f9a6d",
						"name": "svc1",
						"port": 80,
						"protocol": "http",
						"read_timeout": 60000,
						"retries": 5,
						"write_timeout": 60000,
						"tags": [
							"test"
						]
					},
					"old": {
						"connect_timeout": 60000,
						"enabled": true,
						"host": "mockbin.org",
						"id": "9ecf5708-f2f4-444e-a4c7-fcd3a57f9a6d",
						"name": "svc1",
						"port": 80,
						"protocol": "http",
						"read_timeout": 60000,
						"retries": 5,
						"write_timeout": 60000
					}
				}
			}
		],
		"deleting": []
	},
	"summary": {
		"creating": 1,
		"updating": 1,
		"deleting": 0,
		"total": 2
	},
	"warnings": [],
	"errors": []
}

`

	expectedOutputMaskedJSON = `{
	"changes": {
		"creating": [
			{
				"name": "rate-limiting (global)",
				"kind": "plugin",
				"body": {
					"new": {
						"id": "a1368a28-cb5c-4eee-86d8-03a6bdf94b5e",
						"name": "rate-limiting",
						"config": {
							"day": null,
							"error_code": 429,
							"error_message": "API rate limit exceeded",
							"fault_tolerant": true,
							"header_name": null,
							"hide_client_headers": false,
							"hour": null,
							"limit_by": "consumer",
							"minute": 123,
							"month": null,
							"path": null,
							"policy": "local",
							"redis_database": 0,
							"redis_host": null,
							"redis_password": null,
							"redis_port": 6379,
							"redis_server_name": null,
							"redis_ssl": false,
							"redis_ssl_verify": false,
							"redis_timeout": 2000,
							"redis_username": null,
							"second": null,
							"year": null
						},
						"enabled": true,
						"protocols": [
							"grpc",
							"grpcs",
							"http",
							"https"
						]
					},
					"old": null
				}
			}
		],
		"updating": [
			{
				"name": "svc1",
				"kind": "service",
				"body": {
					"new": {
						"connect_timeout": 60000,
						"enabled": true,
						"host": "[masked]",
						"id": "9ecf5708-f2f4-444e-a4c7-fcd3a57f9a6d",
						"name": "svc1",
						"port": 80,
						"protocol": "http",
						"read_timeout": 60000,
						"retries": 5,
						"write_timeout": 60000,
						"tags": [
							"[masked] is an external host. I like [masked]!",
							"foo:foo",
							"baz:[masked]",
							"another:[masked]",
							"bar:[masked]"
						]
					},
					"old": {
						"connect_timeout": 60000,
						"enabled": true,
						"host": "[masked]",
						"id": "9ecf5708-f2f4-444e-a4c7-fcd3a57f9a6d",
						"name": "svc1",
						"port": 80,
						"protocol": "http",
						"read_timeout": 60000,
						"retries": 5,
						"write_timeout": 60000
					}
				}
			}
		],
		"deleting": []
	},
	"summary": {
		"creating": 1,
		"updating": 1,
		"deleting": 0,
		"total": 2
	},
	"warnings": [],
	"errors": []
}

`

	expectedOutputUnMaskedJSON30x = `{
	"changes": {
		"creating": [
			{
				"name": "rate-limiting (global)",
				"kind": "plugin",
				"body": {
					"new": {
						"id": "a1368a28-cb5c-4eee-86d8-03a6bdf94b5e",
						"name": "rate-limiting",
						"config": {
							"day": null,
							"fault_tolerant": true,
							"header_name": null,
							"hide_client_headers": false,
							"hour": null,
							"limit_by": "consumer",
							"minute": 123,
							"month": null,
							"path": null,
							"policy": "local",
							"redis_database": 0,
							"redis_host": null,
							"redis_password": null,
							"redis_port": 6379,
							"redis_server_name": null,
							"redis_ssl": false,
							"redis_ssl_verify": false,
							"redis_timeout": 2000,
							"redis_username": null,
							"second": null,
							"year": null
						},
						"enabled": true,
						"protocols": [
							"grpc",
							"grpcs",
							"http",
							"https"
						]
					},
					"old": null
				}
			}
		],
		"updating": [
			{
				"name": "svc1",
				"kind": "service",
				"body": {
					"new": {
						"connect_timeout": 60000,
						"enabled": true,
						"host": "mockbin.org",
						"id": "9ecf5708-f2f4-444e-a4c7-fcd3a57f9a6d",
						"name": "svc1",
						"port": 80,
						"protocol": "http",
						"read_timeout": 60000,
						"retries": 5,
						"write_timeout": 60000,
						"tags": [
							"test"
						]
					},
					"old": {
						"connect_timeout": 60000,
						"enabled": true,
						"host": "mockbin.org",
						"id": "9ecf5708-f2f4-444e-a4c7-fcd3a57f9a6d",
						"name": "svc1",
						"port": 80,
						"protocol": "http",
						"read_timeout": 60000,
						"retries": 5,
						"write_timeout": 60000
					}
				}
			}
		],
		"deleting": []
	},
	"summary": {
		"creating": 1,
		"updating": 1,
		"deleting": 0,
		"total": 2
	},
	"warnings": [],
	"errors": []
}

`

	expectedOutputMaskedJSON30x = `{
	"changes": {
		"creating": [
			{
				"name": "rate-limiting (global)",
				"kind": "plugin",
				"body": {
					"new": {
						"id": "a1368a28-cb5c-4eee-86d8-03a6bdf94b5e",
						"name": "rate-limiting",
						"config": {
							"day": null,
							"fault_tolerant": true,
							"header_name": null,
							"hide_client_headers": false,
							"hour": null,
							"limit_by": "consumer",
							"minute": 123,
							"month": null,
							"path": null,
							"policy": "local",
							"redis_database": 0,
							"redis_host": null,
							"redis_password": null,
							"redis_port": 6379,
							"redis_server_name": null,
							"redis_ssl": false,
							"redis_ssl_verify": false,
							"redis_timeout": 2000,
							"redis_username": null,
							"second": null,
							"year": null
						},
						"enabled": true,
						"protocols": [
							"grpc",
							"grpcs",
							"http",
							"https"
						]
					},
					"old": null
				}
			}
		],
		"updating": [
			{
				"name": "svc1",
				"kind": "service",
				"body": {
					"new": {
						"connect_timeout": 60000,
						"enabled": true,
						"host": "[masked]",
						"id": "9ecf5708-f2f4-444e-a4c7-fcd3a57f9a6d",
						"name": "svc1",
						"port": 80,
						"protocol": "http",
						"read_timeout": 60000,
						"retries": 5,
						"write_timeout": 60000,
						"tags": [
							"[masked] is an external host. I like [masked]!",
							"foo:foo",
							"baz:[masked]",
							"another:[masked]",
							"bar:[masked]"
						]
					},
					"old": {
						"connect_timeout": 60000,
						"enabled": true,
						"host": "[masked]",
						"id": "9ecf5708-f2f4-444e-a4c7-fcd3a57f9a6d",
						"name": "svc1",
						"port": 80,
						"protocol": "http",
						"read_timeout": 60000,
						"retries": 5,
						"write_timeout": 60000
					}
				}
			}
		],
		"deleting": []
	},
	"summary": {
		"creating": 1,
		"updating": 1,
		"deleting": 0,
		"total": 2
	},
	"warnings": [],
	"errors": []
}

`
)

// test scope:
//   - 1.x
//   - 2.x
func Test_Diff_Workspace_OlderThan3x(t *testing.T) {
	tests := []struct {
		name          string
		stateFile     string
		expectedState utils.KongRawState
	}{
		{
			name:      "diff with not existent workspace doesn't error out",
			stateFile: "testdata/diff/001-not-existing-workspace/kong.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", "<3.0.0")
			setup(t)

			_, err := diff(tc.stateFile)
			require.NoError(t, err)
		})
	}
}

// test scope:
//   - 3.x
func Test_Diff_Workspace_NewerThan3x(t *testing.T) {
	tests := []struct {
		name          string
		stateFile     string
		expectedState utils.KongRawState
	}{
		{
			name:      "diff with not existent workspace doesn't error out",
			stateFile: "testdata/diff/001-not-existing-workspace/kong3x.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.0.0")
			setup(t)

			_, err := diff(tc.stateFile)
			require.NoError(t, err)
		})
	}
}

// test scope:
//   - 2.8.0
func Test_Diff_Masked_OlderThan3x(t *testing.T) {
	tests := []struct {
		name             string
		initialStateFile string
		stateFile        string
		expectedState    utils.KongRawState
		envVars          map[string]string
	}{
		{
			name:             "env variable are masked",
			initialStateFile: "testdata/diff/002-mask/initial.yaml",
			stateFile:        "testdata/diff/002-mask/kong.yaml",
			envVars:          diffEnvVars,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}
			runWhen(t, "kong", "==2.8.0")
			setup(t)

			// initialize state
			require.NoError(t, sync(context.Background(), tc.initialStateFile))

			out, err := diff(tc.stateFile)
			require.NoError(t, err)
			assert.Equal(t, expectedOutputMasked, out)
		})
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}
			runWhen(t, "kong", "==2.8.0")
			setup(t)

			// initialize state
			require.NoError(t, sync(context.Background(), tc.initialStateFile))

			out, err := diff(tc.stateFile, "--json-output")
			require.NoError(t, err)
			assert.JSONEq(t, expectedOutputMaskedJSON, out)
		})
	}
}

// test scope:
//   - 3.x
func Test_Diff_Masked_NewerThan3x(t *testing.T) {
	tests := []struct {
		name             string
		initialStateFile string
		stateFile        string
		expectedState    utils.KongRawState
		envVars          map[string]string
	}{
		{
			name:             "env variable are masked",
			initialStateFile: "testdata/diff/002-mask/initial3x.yaml",
			stateFile:        "testdata/diff/002-mask/kong3x.yaml",
			envVars:          diffEnvVars,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}
			runWhen(t, "kong", ">=3.0.0")
			setup(t)

			// initialize state
			require.NoError(t, sync(context.Background(), tc.initialStateFile))

			out, err := diff(tc.stateFile)
			require.NoError(t, err)
			assert.Equal(t, expectedOutputMasked, out)
		})
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}
			runWhen(t, "kong", ">=3.0.0 <3.1.0")
			setup(t)

			// initialize state
			require.NoError(t, sync(context.Background(), tc.initialStateFile))

			out, err := diff(tc.stateFile, "--json-output")
			require.NoError(t, err)
			assert.JSONEq(t, expectedOutputMaskedJSON30x, out)
		})
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}
			runWhen(t, "kong", ">=3.1.0 <3.4.0")
			setup(t)

			// initialize state
			require.NoError(t, sync(context.Background(), tc.initialStateFile))

			out, err := diff(tc.stateFile, "--json-output")
			require.NoError(t, err)
			assert.JSONEq(t, expectedOutputMaskedJSON, out)
		})
	}
}

// test scope:
//   - 2.8.0
func Test_Diff_Unmasked_OlderThan3x(t *testing.T) {
	tests := []struct {
		name             string
		initialStateFile string
		stateFile        string
		expectedState    utils.KongRawState
		envVars          map[string]string
	}{
		{
			name:             "env variable are unmasked",
			initialStateFile: "testdata/diff/003-unmask/initial.yaml",
			stateFile:        "testdata/diff/003-unmask/kong.yaml",
			envVars:          diffEnvVars,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}
			runWhen(t, "kong", "==2.8.0")
			setup(t)

			// initialize state
			require.NoError(t, sync(context.Background(), tc.initialStateFile))

			out, err := diff(tc.stateFile, "--no-mask-deck-env-vars-value")
			require.NoError(t, err)
			assert.Equal(t, expectedOutputUnMasked, out)
		})
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}
			runWhen(t, "kong", "==2.8.0")
			setup(t)

			// initialize state
			require.NoError(t, sync(context.Background(), tc.initialStateFile))

			out, err := diff(tc.stateFile, "--no-mask-deck-env-vars-value", "--json-output")
			require.NoError(t, err)
			assert.JSONEq(t, expectedOutputUnMaskedJSON, out)
		})
	}
}

// test scope:
//   - 3.x
func Test_Diff_Unmasked_NewerThan3x(t *testing.T) {
	tests := []struct {
		name             string
		initialStateFile string
		stateFile        string
		expectedState    utils.KongRawState
		envVars          map[string]string
	}{
		{
			name:             "env variable are unmasked",
			initialStateFile: "testdata/diff/003-unmask/initial3x.yaml",
			stateFile:        "testdata/diff/003-unmask/kong3x.yaml",
			envVars:          diffEnvVars,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}
			runWhen(t, "kong", ">=3.0.0")
			setup(t)

			// initialize state
			require.NoError(t, sync(context.Background(), tc.initialStateFile))

			out, err := diff(tc.stateFile, "--no-mask-deck-env-vars-value")
			require.NoError(t, err)
			assert.Equal(t, expectedOutputUnMasked, out)
		})
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}
			runWhen(t, "kong", ">=3.0.0 <3.1.0")
			setup(t)

			// initialize state
			require.NoError(t, sync(context.Background(), tc.initialStateFile))

			out, err := diff(tc.stateFile, "--no-mask-deck-env-vars-value", "--json-output")
			require.NoError(t, err)
			assert.JSONEq(t, expectedOutputUnMaskedJSON30x, out)
		})
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}
			runWhen(t, "kong", ">=3.1.0 <3.4.0")
			setup(t)

			// initialize state
			require.NoError(t, sync(context.Background(), tc.initialStateFile))

			out, err := diff(tc.stateFile, "--no-mask-deck-env-vars-value", "--json-output")
			require.NoError(t, err)
			assert.JSONEq(t, expectedOutputUnMaskedJSON, out)
		})
	}
}

// test scope:
//   - 3.5
func Test_Diff_NoDiffUnorderedArray(t *testing.T) {
	tests := []struct {
		name      string
		stateFile string
		runWhen   string
	}{
		{
			name:      "no diffs with unordered arrays >=3.5.0 <3.8.0",
			stateFile: "testdata/diff/004-no-diff-plugin/kong.yaml",
			runWhen:   ">=3.5.0 <3.8.0",
		},
		{
			name:      "no diffs with unordered arrays >=3.8.0",
			stateFile: "testdata/diff/004-no-diff-plugin/kong.yaml",
			runWhen:   ">=3.8.0",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "enterprise", tc.runWhen)
			setup(t)

			// test that the diff command does not return any changes when
			// array fields are not sorted.
			require.NoError(t, sync(context.Background(), tc.stateFile, "--timeout", "60"))

			out, err := diff(tc.stateFile)
			require.NoError(t, err)
			assert.Equal(t, emptyOutput, out)
			reset(t)
		})
	}
}

func Test_Diff_Plugin_Protocol_Order_Change(t *testing.T) {
	setup(t)
	ctx := context.Background()

	tests := []struct {
		name             string
		initialStateFile string
		stateFile        string
		expectedDiff     string
	}{
		{
			name:             "syncing and then diffing same file should not show false diff",
			initialStateFile: "testdata/diff/004-no-diff-plugin/initial-ip-restriction.yaml",
			stateFile:        "testdata/diff/004-no-diff-plugin/initial-ip-restriction.yaml",
			expectedDiff:     emptyOutput,
		},
		{
			name:             "changing protocol order should not show diff",
			initialStateFile: "testdata/diff/004-no-diff-plugin/protocol-initial-order-plugins.yaml",
			stateFile:        "testdata/diff/004-no-diff-plugin/protocol-reordered-plugins.yaml",
			expectedDiff:     emptyOutput,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// initialize state
			require.NoError(t, sync(ctx, tc.initialStateFile))

			out, err := diff(tc.stateFile)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedDiff, out)
		})
	}
}

// test scope:
//   - 3.5
func Test_Diff_NoDiffCompressedTarget(t *testing.T) {
	runWhen(t, "kong", ">=3.5.0")
	setup(t)

	// test that the diff command does not return any changes when
	// target is a compressed IPv6.
	stateFile := "testdata/diff/005-no-diff-target/kong.yaml"
	require.NoError(t, sync(context.Background(), stateFile))

	out, err := diff(stateFile)
	require.NoError(t, err)
	assert.Equal(t, emptyOutput, out)
	reset(t)
}
