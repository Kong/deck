//go:build integration

package integration

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/kong/deck/sanitize"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Dump_SelectTags_30(t *testing.T) {
	tests := []struct {
		name         string
		stateFile    string
		expectedFile string
	}{
		{
			name:         "dump with select-tags",
			stateFile:    "testdata/dump/001-entities-with-tags/kong.yaml",
			expectedFile: "testdata/dump/001-entities-with-tags/expected30.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.0.0 <3.1.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.stateFile))

			output, err := dump(
				"--select-tag", "managed-by-deck",
				"--select-tag", "org-unit-42",
				"-o", "-",
			)
			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

func Test_Dump_SelectTags_3x(t *testing.T) {
	tests := []struct {
		name         string
		stateFile    string
		expectedFile string
		runWhen      string
	}{
		{
			name:         "dump with select-tags >=3.1.0 <3.8.0",
			stateFile:    "testdata/dump/001-entities-with-tags/kong.yaml",
			expectedFile: "testdata/dump/001-entities-with-tags/expected.yaml",
			runWhen:      ">=3.1.0 <3.8.0",
		},
		{
			name:         "dump with select-tags >=3.8.0 <3.10.0",
			stateFile:    "testdata/dump/001-entities-with-tags/kong.yaml",
			expectedFile: "testdata/dump/001-entities-with-tags/expected381.yaml",
			runWhen:      ">=3.8.0 <3.10.0",
		},
		{
			name:         "dump with select-tags >=3.10.0 <3.11.0",
			stateFile:    "testdata/dump/001-entities-with-tags/kong.yaml",
			expectedFile: "testdata/dump/001-entities-with-tags/expected310.yaml",
			runWhen:      ">=3.10.0 <3.11.0",
		},
		{
			name:         "dump with select-tags >=3.11.0",
			stateFile:    "testdata/dump/001-entities-with-tags/kong.yaml",
			expectedFile: "testdata/dump/001-entities-with-tags/expected311.yaml",
			runWhen:      ">=3.11.0 <3.12.0",
		},
		{
			name:         "dump with select-tags >=3.12.0",
			stateFile:    "testdata/dump/001-entities-with-tags/kong.yaml",
			expectedFile: "testdata/dump/001-entities-with-tags/expected312.yaml",
			runWhen:      ">=3.12.0",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", tc.runWhen)
			setup(t)

			require.NoError(t, sync(context.Background(), tc.stateFile))

			output, err := dump(
				"--select-tag", "managed-by-deck",
				"--select-tag", "org-unit-42",
				"-o", "-",
			)
			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

func Test_Dump_SkipConsumers(t *testing.T) {
	tests := []struct {
		name          string
		stateFile     string
		expectedFile  string
		skipConsumers bool
		runWhen       func(t *testing.T)
	}{
		{
			name:          "3.2 & 3.3 dump with skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected.yaml",
			skipConsumers: true,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.2.0 <3.4.0") },
		},
		{
			name:          "3.2 & 3.3 dump with no skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected-no-skip.yaml",
			skipConsumers: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.2.0 <3.4.0") },
		},
		{
			name:          "3.4 dump with skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong34.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected.yaml",
			skipConsumers: true,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.4.0 <3.5.0") },
		},
		{
			name:          "3.4 dump with no skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong34.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected-no-skip-34.yaml",
			skipConsumers: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.4.0 <3.5.0") },
		},
		{
			name:          "3.5 dump with skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong34.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected.yaml",
			skipConsumers: true,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.5.0") },
		},
		{
			name:          "3.5 dump with no skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong34.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected-no-skip-35.yaml",
			skipConsumers: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.5.0 <3.8.0") },
		},
		{
			name:          "3.8.0 dump with no skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong34.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected-no-skip-381.yaml",
			skipConsumers: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.8.0 <3.9.0") },
		},
		{
			name:          "3.9.0 dump with no skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong34.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected-no-skip-39.yaml",
			skipConsumers: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.9.0 <3.10.0") },
		},
		{
			name:          "3.10.0 dump with no skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong34.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected-no-skip-310.yaml",
			skipConsumers: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.10.0 <3.12.0") },
		},
		{
			name:          ">=3.12.0 dump with no skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong34.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected-no-skip-312.yaml",
			skipConsumers: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.12.0 <3.13.0") },
		},
		{
			name:          ">=3.13.0 dump with no skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong34.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected-no-skip-313.yaml",
			skipConsumers: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.13.0 <3.14.0") },
		},
		{
			name:          ">=3.14.0 dump with no skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong34.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected-no-skip-314.yaml",
			skipConsumers: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.14.0") },
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.runWhen(t)
			setup(t)

			require.NoError(t, sync(context.Background(), tc.stateFile))

			var (
				output string
				err    error
			)
			if tc.skipConsumers {
				output, err = dump(
					"--skip-consumers",
					"-o", "-",
				)
			} else {
				output, err = dump(
					"-o", "-",
				)
			}
			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

func Test_Dump_SkipConsumers_Konnect(t *testing.T) {
	tests := []struct {
		name          string
		stateFile     string
		expectedFile  string
		skipConsumers bool
	}{
		{
			name:          "dump with skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong34.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected_konnect.yaml",
			skipConsumers: true,
		},
		{
			name:          "dump with no skip-consumers",
			stateFile:     "testdata/dump/002-skip-consumers/kong34.yaml",
			expectedFile:  "testdata/dump/002-skip-consumers/expected-no-skip_konnect.yaml",
			skipConsumers: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhenKonnect(t)
			setup(t)

			require.NoError(t, sync(context.Background(), tc.stateFile))

			var (
				output string
				err    error
			)
			if tc.skipConsumers {
				output, err = dump(
					"--skip-consumers",
					"-o", "-",
				)
			} else {
				output, err = dump(
					"-o", "-",
				)
			}
			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

func Test_Dump_KonnectRename(t *testing.T) {
	tests := []struct {
		name         string
		stateFile    string
		expectedFile string
		flags        []string
	}{
		{
			name:         "dump with konnect-control-plane-name",
			stateFile:    "testdata/sync/026-konnect-rename/konnect_test_cp.yaml",
			expectedFile: "testdata/sync/026-konnect-rename/konnect_test_cp.yaml",
			flags:        []string{"--konnect-control-plane-name", "test"},
		},
		{
			name:         "dump with konnect-runtime-group-name",
			stateFile:    "testdata/sync/026-konnect-rename/konnect_test_rg.yaml",
			expectedFile: "testdata/sync/026-konnect-rename/konnect_test_cp.yaml",
			flags:        []string{"--konnect-runtime-group-name", "test"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				reset(t, tc.flags...)
			})
			runWhenKonnect(t)
			setup(t)

			require.NoError(t, sync(context.Background(), tc.stateFile))

			var (
				output string
				err    error
			)
			flags := make([]string, 0, 3+len(tc.flags))
			flags = append(flags, "-o", "-", "--with-id")
			flags = append(flags, tc.flags...)
			output, err = dump(flags...)

			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

func Test_Dump_ConsumerGroupConsumersWithCustomID(t *testing.T) {
	runWhen(t, "enterprise", ">=3.0.0")
	setup(t)

	require.NoError(t, sync(context.Background(), "testdata/sync/028-consumer-group-consumers-custom_id/kong.yaml"))

	var output string
	flags := []string{"-o", "-", "--with-id"}
	output, err := dump(flags...)
	require.NoError(t, err)

	expected, err := readFile("testdata/sync/028-consumer-group-consumers-custom_id/kong.yaml")
	require.NoError(t, err)
	assert.Equal(t, expected, output)
}

func Test_Dump_ConsumerGroupConsumersWithCustomID_Konnect(t *testing.T) {
	runWhen(t, "konnect", "")
	setup(t)

	require.NoError(t, sync(context.Background(), "testdata/sync/028-consumer-group-consumers-custom_id/kong.yaml"))

	var output string
	flags := []string{"-o", "-", "--with-id"}
	output, err := dump(flags...)
	require.NoError(t, err)

	expected, err := readFile("testdata/dump/003-consumer-group-consumers-custom_id/konnect.yaml")
	require.NoError(t, err)
	assert.Equal(t, expected, output)
}

func Test_Dump_FilterChains(t *testing.T) {
	t.Skip("Skipping test till wasm/filter-chains issue is not resolved at gateway level")
	runWhen(t, "kong", ">=3.4.0")
	setup(t)

	tests := []struct {
		version  string
		input    string
		expected string
	}{
		{
			version:  "<3.5.0",
			input:    "testdata/dump/004-filter-chains/kong-3.4.x.yaml",
			expected: "testdata/dump/004-filter-chains/expected-3.4.x.yaml",
		},
		{
			version:  ">=3.5.0",
			input:    "testdata/dump/004-filter-chains/kong.yaml",
			expected: "testdata/dump/004-filter-chains/expected.yaml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.version, func(t *testing.T) {
			runWhen(t, "kong", tc.version)
			require.NoError(t, sync(context.Background(), tc.input))

			var output string
			flags := []string{"-o", "-"}
			output, err := dump(flags...)
			require.NoError(t, err)

			expected, err := readFile(tc.expected)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

func Test_SkipConsumersWithConsumerGroups(t *testing.T) {
	tests := []struct {
		name                            string
		stateFile                       string
		expectedFile                    string
		errorExpected                   bool
		errorString                     string
		skipConsumersWithConsumerGroups bool
		runWhen                         func(t *testing.T)
	}{
		{
			name:                            "dump with flag --skip-consumers-with-consumer-groups set: <3.0.0 ",
			stateFile:                       "testdata/dump/004-skip-consumers-with-consumer-groups/kong.yaml",
			expectedFile:                    "testdata/dump/004-skip-consumers-with-consumer-groups/expected-with-flag_1.yaml",
			skipConsumersWithConsumerGroups: true,
			runWhen:                         func(t *testing.T) { runWhen(t, "enterprise", "<3.0.0") },
		},
		{
			name:                            "dump with flag --skip-consumers-with-consumer-groups not set: <3.0.0 ",
			stateFile:                       "testdata/dump/004-skip-consumers-with-consumer-groups/kong.yaml",
			expectedFile:                    "testdata/dump/004-skip-consumers-with-consumer-groups/expected-no-flag_1.yaml",
			skipConsumersWithConsumerGroups: false,
			runWhen:                         func(t *testing.T) { runWhen(t, "enterprise", "<3.0.0") },
		},
		{
			name:                            "dump with flag --skip-consumers-with-consumer-groups set: <3.9.0 ",
			stateFile:                       "testdata/dump/004-skip-consumers-with-consumer-groups/kong3x.yaml",
			expectedFile:                    "testdata/dump/004-skip-consumers-with-consumer-groups/expected-with-flag.yaml",
			skipConsumersWithConsumerGroups: true,
			runWhen:                         func(t *testing.T) { runWhen(t, "enterprise", ">=3.0.0 <3.9.0") },
		},
		{
			name:                            "dump with flag --skip-consumers-with-consumer-groups not set: <3.9.0 ",
			stateFile:                       "testdata/dump/004-skip-consumers-with-consumer-groups/kong3x.yaml",
			expectedFile:                    "testdata/dump/004-skip-consumers-with-consumer-groups/expected-no-flag.yaml",
			skipConsumersWithConsumerGroups: false,
			runWhen:                         func(t *testing.T) { runWhen(t, "enterprise", ">=3.0.0 <3.9.0") },
		},
		{
			name:                            "dump with flag --skip-consumers-with-consumer-groups set: >=3.9.0 ",
			stateFile:                       "testdata/dump/004-skip-consumers-with-consumer-groups/kong3x.yaml",
			expectedFile:                    "testdata/dump/004-skip-consumers-with-consumer-groups/expected-with-flag.yaml",
			skipConsumersWithConsumerGroups: true,
			runWhen:                         func(t *testing.T) { runWhen(t, "enterprise", ">=3.9.0") },
		},
		{
			name:                            "dump with flag --skip-consumers-with-consumer-groups not set: >=3.9.0 ",
			stateFile:                       "testdata/dump/004-skip-consumers-with-consumer-groups/kong3x.yaml",
			expectedFile:                    "testdata/dump/004-skip-consumers-with-consumer-groups/expected-no-flag.yaml",
			skipConsumersWithConsumerGroups: false,
			runWhen:                         func(t *testing.T) { runWhen(t, "enterprise", ">=3.9.0") },
		},
		{
			name:                            "dump with flag --skip-consumers-with-consumer-groups set: Konnect ",
			stateFile:                       "testdata/dump/004-skip-consumers-with-consumer-groups/kong3x.yaml",
			skipConsumersWithConsumerGroups: true,
			runWhen:                         func(t *testing.T) { runWhenKonnect(t) },
			errorExpected:                   true,
			errorString:                     "the flag --skip-consumers-with-consumer-groups can not be used with Konnect",
		},
		{
			name:                            "dump with flag --skip-consumers-with-consumer-groups not set: Konnect ",
			stateFile:                       "testdata/dump/004-skip-consumers-with-consumer-groups/kong3x.yaml",
			expectedFile:                    "testdata/dump/004-skip-consumers-with-consumer-groups/expected-konnect.yaml",
			skipConsumersWithConsumerGroups: false,
			runWhen:                         func(t *testing.T) { runWhenKonnect(t) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.runWhen(t)
			setup(t)

			require.NoError(t, sync(context.Background(), tc.stateFile))

			var (
				output string
				err    error
			)
			if tc.skipConsumersWithConsumerGroups {
				output, err = dump(
					"--skip-consumers-with-consumer-groups",
					"-o", "-",
				)
			} else {
				output, err = dump(
					"-o", "-",
				)
			}

			if tc.errorExpected {
				assert.Equal(t, err.Error(), tc.errorString)
				return
			}

			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

func Test_Dump_ConsumerGroupPlugin_PolicyOverrides(t *testing.T) {
	tests := []struct {
		name          string
		stateFile     string
		expectedFile  string
		errorExpected bool
		errorString   string
		runWhen       func(t *testing.T)
	}{
		{
			name:          "dump with flag --consumer-group-policy-overrides set: >=3.4.0 <3.8.0",
			stateFile:     "testdata/sync/037-consumer-group-policy-overrides/kong34x-no-info.yaml",
			expectedFile:  "testdata/sync/037-consumer-group-policy-overrides/kong34x.yaml",
			errorExpected: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.4.0 <3.8.0") },
		},
		{
			name:          "dump with flag --consumer-group-policy-overrides set: >=3.8.0 <3.9.0",
			stateFile:     "testdata/sync/037-consumer-group-policy-overrides/kong38x-no-info.yaml",
			expectedFile:  "testdata/sync/037-consumer-group-policy-overrides/kong38x.yaml",
			errorExpected: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.8.0 <3.9.0") },
		},
		{
			name:          "dump with flag --consumer-group-policy-overrides set: >=3.9.0",
			stateFile:     "testdata/sync/037-consumer-group-policy-overrides/kong39x-no-info.yaml",
			expectedFile:  "testdata/sync/037-consumer-group-policy-overrides/kong39x.yaml",
			errorExpected: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.9.0 <3.12.0") },
		},
		{
			name:          "dump with flag --consumer-group-policy-overrides set: >=3.12.0",
			stateFile:     "testdata/sync/037-consumer-group-policy-overrides/kong39x-no-info.yaml",
			expectedFile:  "testdata/sync/037-consumer-group-policy-overrides/kong312x.yaml",
			errorExpected: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.12.0 <3.13.0") },
		},
		{
			name:          "dump with flag --consumer-group-policy-overrides set: >=3.13.0",
			stateFile:     "testdata/sync/037-consumer-group-policy-overrides/kong39x-no-info.yaml",
			expectedFile:  "testdata/sync/037-consumer-group-policy-overrides/kong313x.yaml",
			errorExpected: false,
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.13.0") },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.runWhen(t)
			setup(t)

			require.NoError(t, sync(context.Background(), tc.stateFile, "--consumer-group-policy-overrides"))

			var (
				output string
				err    error
			)

			output, err = dump(
				"--consumer-group-policy-overrides",
				"-o", "-",
			)

			if tc.errorExpected {
				assert.Equal(t, err.Error(), tc.errorString)
				return
			}

			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

// test scope:
//
// - >=3.1.0
func Test_Dump_KeysAndKeySets(t *testing.T) {
	runWhen(t, "kong", ">=3.1.0")
	setup(t)
	ctx := context.Background()

	tests := []struct {
		name         string
		stateFile    string
		expectedFile string
	}{
		{
			name:         "dump keys and key-sets - jwk keys",
			stateFile:    "testdata/dump/007-keys-and-key_sets/kong-jwk.yaml",
			expectedFile: "testdata/dump/007-keys-and-key_sets/kong-jwk.yaml",
		},
		{
			name:         "dump keys and key-sets - pem keys",
			stateFile:    "testdata/dump/007-keys-and-key_sets/kong-pem.yaml",
			expectedFile: "testdata/dump/007-keys-and-key_sets/kong-pem.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, sync(ctx, tc.stateFile))

			output, err := dump("-o", "-", "--with-id")
			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

// test scope:
//
// - konnect
func Test_Dump_KeysAndKeySets_Konnect(t *testing.T) {
	runWhenKonnect(t)
	setup(t)
	ctx := context.Background()

	tests := []struct {
		name         string
		stateFile    string
		expectedFile string
	}{
		{
			name:         "dump keys and key-sets - jwk keys",
			stateFile:    "testdata/dump/007-keys-and-key_sets/kong-jwk.yaml",
			expectedFile: "testdata/dump/007-keys-and-key_sets/konnect-jwk.yaml",
		},
		{
			name:         "dump keys and key-sets - pem keys",
			stateFile:    "testdata/dump/007-keys-and-key_sets/kong-pem.yaml",
			expectedFile: "testdata/dump/007-keys-and-key_sets/konnect-pem.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, sync(ctx, tc.stateFile))

			output, err := dump("-o", "-", "--with-id")
			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

// test scope:
//
// - >=3.10.0
func Test_Dump_Deterministic_Sanitizer(t *testing.T) {
	runWhen(t, "enterprise", ">=3.10.0")
	setup(t)
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name         string
		stateFile    string
		expectedFile string
	}{
		{
			name:         "dump with sanitizer",
			stateFile:    "testdata/dump/008-sanitizer/kong.yaml",
			expectedFile: "testdata/dump/008-sanitizer/expected.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, sync(context.Background(), tc.stateFile))

			fileContent, err := file.GetContentFromFiles([]string{tc.stateFile}, false)
			require.NoError(t, err)

			sanitizer := sanitize.NewSanitizer(&sanitize.SanitizerOptions{
				Ctx:     ctx,
				Client:  client,
				Content: fileContent.DeepCopy(),
				Salt:    "deterministic-test-salt-12345",
			})

			sanitizedOutput, err := sanitizer.Sanitize()
			require.NoError(t, err)

			// capture command output to be used during tests
			rescueStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			err = file.WriteContentToFile(sanitizedOutput, "-", file.YAML)
			require.NoError(t, err)
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = rescueStdout

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			require.Equal(t, expected, stripansi.Strip(string(out)))

			// validate file content
			validateOpts := []string{tc.expectedFile}
			err = validate(ONLINE, validateOpts...)
			require.NoError(t, err)
		})
	}
}

func Test_Dump_Sanitize(t *testing.T) {
	ctx := context.Background()
	testSalt := "test-salt-123"

	tests := []struct {
		name         string
		stateFile    string
		expectedFile string
		selectTags   []string
		runWhen      func(t *testing.T)
	}{
		{
			name:         "dump sanitized services and routes",
			stateFile:    "testdata/dump/008-sanitizer/services-routes.yaml",
			expectedFile: "testdata/dump/008-sanitizer/services-routes.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "kong", ">=3.0.0 <3.14.0") },
		},
		{
			name:         "dump sanitized services and routes >=3.14.0",
			stateFile:    "testdata/dump/008-sanitizer/services-routes.yaml",
			expectedFile: "testdata/dump/008-sanitizer/services-routes.expected314.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "kong", ">=3.14.0") },
		},
		{
			name:         "dump sanitized consumers, consumer-groups and consumer-group-plugins",
			stateFile:    "testdata/dump/008-sanitizer/consumergroup-plugins.yaml",
			expectedFile: "testdata/dump/008-sanitizer/consumergroup-plugins.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", "3.4.0") },
		},
		{
			name:         "dump sanitized consumers, consumer-groups and consumer-group-plugins >=3.6.0",
			stateFile:    "testdata/dump/008-sanitizer/consumergroup-plugins36.yaml",
			expectedFile: "testdata/dump/008-sanitizer/consumergroup-plugins36.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.6.0 <3.13.0") },
		},
		{
			name:         "dump sanitized consumers, consumer-groups and consumer-group-plugins >=3.13.0",
			stateFile:    "testdata/dump/008-sanitizer/consumergroup-plugins36.yaml",
			expectedFile: "testdata/dump/008-sanitizer/consumergroup-plugins313.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.13.0 <3.14.0") },
		},
		{
			name:         "dump sanitized consumers, consumer-groups and consumer-group-plugins >=3.14.0",
			stateFile:    "testdata/dump/008-sanitizer/consumergroup-plugins36.yaml",
			expectedFile: "testdata/dump/008-sanitizer/consumergroup-plugins314.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.14.0") },
		},
		{
			name:         "dump sanitize with select-tags",
			stateFile:    "testdata/dump/008-sanitizer/consumergroup-plugins36.yaml",
			expectedFile: "testdata/dump/008-sanitizer/select-tags.expected.yaml",
			selectTags:   []string{"tag1", "tag2"},
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.6.0") },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.runWhen(t)
			setup(t)
			require.NoError(t, sync(ctx, tc.stateFile))

			dumpArgs := []string{"-o", "-", "--sanitize", "--sanitization-salt", testSalt}
			if len(tc.selectTags) > 0 {
				dumpArgs = append(dumpArgs, "--select-tag", strings.Join(tc.selectTags, ","))
			}

			// checking that the sanitizer is working correctly
			output, err := dump(dumpArgs...)
			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			require.Equal(t, expected, output)

			// validate file content
			validateOpts := []string{tc.expectedFile}
			err = validate(ONLINE, validateOpts...)
			require.NoError(t, err)

			// re-syncing to ensure the sanitized content is valid
			reset(t)
			require.NoError(t, sync(ctx, tc.expectedFile))
		})
	}
}

func Test_Dump_Sanitize_Special_Entities(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		stateFile      string
		runWhen        func(t *testing.T)
		skipValidation bool // some
	}{
		{
			name:      "dump sanitize keys - jwk",
			stateFile: "testdata/dump/007-keys-and-key_sets/kong-jwk.yaml",
			runWhen:   func(t *testing.T) { runWhen(t, "kong", ">=3.1.0") },
		},
		{
			name:      "dump sanitize keys - pem",
			stateFile: "testdata/dump/007-keys-and-key_sets/kong-pem.yaml",
			runWhen:   func(t *testing.T) { runWhen(t, "kong", ">=3.1.0") },
		},
		{
			name:      "dump sanitize certificates",
			stateFile: "testdata/dump/008-sanitizer/cert.yaml",
			runWhen:   func(t *testing.T) { runWhen(t, "kong", ">=2.8.0") },
		},
		{
			name:      "dump sanitize ca-certificates",
			stateFile: "testdata/dump/008-sanitizer/ca-cert.yaml",
			runWhen:   func(t *testing.T) { runWhen(t, "kong", ">=2.8.0") },
		},
		{
			name:           "dump sanitize env vault and vault references",
			stateFile:      "testdata/dump/008-sanitizer/env-vault.yaml",
			runWhen:        func(t *testing.T) { runWhen(t, "enterprise", "3.4.0") },
			skipValidation: true, // env vault validation endpoint not available in 3.4.0
		},
		{
			name:           "dump sanitize env vault and vault references",
			stateFile:      "testdata/dump/008-sanitizer/env-vault.yaml",
			runWhen:        func(t *testing.T) { runWhen(t, "enterprise", ">=3.5.0") },
			skipValidation: true, // env vault validation is flaky in GH actions, skipping for now
		},
		{
			name:           "dump sanitize vault config",
			stateFile:      "testdata/dump/008-sanitizer/vault-configs.yaml",
			runWhen:        func(t *testing.T) { runWhen(t, "enterprise", ">=3.0.0 <3.7.0") },
			skipValidation: true, // vault validation endpoint (for vaults other than env) is not available in prior to 3.7.0
		},
		{
			name:      "dump sanitize vault config",
			stateFile: "testdata/dump/008-sanitizer/vault-configs.yaml",
			runWhen:   func(t *testing.T) { runWhen(t, "enterprise", ">=3.7.0") },
		},
		{
			name:      "dump sanitize vault config >=3.11.0",
			stateFile: "testdata/dump/008-sanitizer/vault-configs311.yaml",
			runWhen:   func(t *testing.T) { runWhen(t, "enterprise", ">=3.11.0") },
		},
		{
			name:      "dump sanitize route expressions",
			stateFile: "testdata/dump/008-sanitizer/route-expressions.yaml",
			runWhen:   func(t *testing.T) { runWhenExpressions(t, ">=3.0.0") },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.runWhen(t)
			setup(t)
			require.NoError(t, sync(ctx, tc.stateFile))

			input, err := readFile(tc.stateFile)
			require.NoError(t, err)

			output, err := dump("-o", "-", "--sanitize")
			require.NoError(t, err)

			require.NotEqual(t, input, output)

			// creating a temp file to write the sanitized output
			tmpFile, err := os.CreateTemp("", "test-output-*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(output)
			require.NoError(t, err)
			tmpFile.Close()

			if !tc.skipValidation {
				// validate file content
				validateOpts := []string{tmpFile.Name()}
				err = validate(ONLINE, validateOpts...)
				require.NoError(t, err)
			}

			// re-syncing to ensure the sanitized content is valid
			reset(t)
			require.NoError(t, sync(ctx, tmpFile.Name()))
		})
	}
}

// test scope:
//
// - konnect
func Test_Dump_BasicAuth_SkipHash(t *testing.T) {
	setDefaultKonnectControlPlane(t)
	runWhenKonnect(t)
	setup(t)

	ctx := context.Background()
	require.NoError(t, sync(ctx, "testdata/sync/047-basic-auth-skip-hash/kong.yaml", "--skip-hash-for-basic-auth"))

	output, err := dump("-o", "-")
	require.NoError(t, err)

	expected, err := readFile("testdata/sync/047-basic-auth-skip-hash/expected-dump.yaml")
	require.NoError(t, err)
	assert.Equal(t, expected, output)
}

// test scope:
//
// - konnect
func Test_Dump_SkipDefaults_Konnect(t *testing.T) {
	setDefaultKonnectControlPlane(t)
	runWhenKonnect(t)
	setup(t)

	ctx := context.Background()

	tests := []struct {
		name         string
		stateFile    string
		expectedFile string
	}{
		{
			name:         "dump skip-defaults: service, routes, service-scoped plugins",
			stateFile:    "testdata/dump/009-skip-defaults/konnect/service-scoped.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/konnect/service-scoped.expected.yaml",
		},
		{
			name:         "dump skip-defaults: consumers, consumer-groups, consumer-group scoped plugins",
			stateFile:    "testdata/dump/009-skip-defaults/konnect/consumer-group-scoped.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/konnect/consumer-group-scoped.expected.yaml",
		},
		{
			name:         "dump skip-defaults: plugins, partials (rla)",
			stateFile:    "testdata/dump/009-skip-defaults/konnect/plugin-partial.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/konnect/plugin-partial.expected.yaml",
		},
		{
			name:         "dump skip-defaults: plugins, partials (openid-connect)",
			stateFile:    "testdata/dump/009-skip-defaults/konnect/plugin-partial-2.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/konnect/plugin-partial-2.expected.yaml",
		},
		{
			name:         "dump skip-defaults: plugin  (rate-limiting)",
			stateFile:    "testdata/dump/009-skip-defaults/konnect/plugin.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/konnect/plugin.expected.yaml",
		},
		{
			name:         "dump skip-defaults: non-default shorthand fields",
			stateFile:    "testdata/dump/009-skip-defaults/konnect/plugin-shorthand-non-default.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/konnect/plugin-shorthand-non-default.expected.yaml",
		},
		{
			name:         "dump skip-defaults: vaults",
			stateFile:    "testdata/dump/009-skip-defaults/konnect/vaults.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/konnect/vaults.expected.yaml",
		},
		{
			name:         "dump skip-defaults: credentials",
			stateFile:    "testdata/dump/009-skip-defaults/konnect/credentials.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/konnect/credentials.expected.yaml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			require.NoError(t, sync(ctx, tc.stateFile))

			dumpArgs := []string{"-o", "-", "--skip-defaults"}
			output, err := dump(dumpArgs...)
			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)

			// ensure that dump can sync back without errors
			require.NoError(t, sync(ctx, tc.expectedFile))

			// dump again
			output, err = dump(dumpArgs...)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

// test scope:
//
// - gateway enterprise
// - 3.4 and 3.10+
func Test_Dump_SkipDefaults(t *testing.T) {
	setup(t)
	ctx := context.Background()

	tests := []struct {
		name         string
		stateFile    string
		expectedFile string
		runWhen      func(t *testing.T)
	}{
		{
			name:         "dump skip-defaults: service, routes, service-scoped plugins 3.4",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.4/service-scoped.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.4/service-scoped.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.4.0 <3.5.0") },
		},
		{
			name:         "dump skip-defaults: service, routes, service-scoped plugins 3.10+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.10+/service-scoped.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.10+/service-scoped.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.10.0 <3.14.0") },
		},
		{
			name:         "dump skip-defaults: service, routes, service-scoped plugins 3.14+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.14+/service-scoped.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.14+/service-scoped.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.14.0") },
		},
		{
			name:         "dump skip-defaults: expression routes",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.4/expression-routes.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.4/expression-routes.expected.yaml",
			runWhen:      func(t *testing.T) { runWhenExpressions(t, ">=3.4.0 <3.14.0") },
		},
		{
			name:         "dump skip-defaults: expression routes >=3.14.0",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.14+/expression-routes.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.14+/expression-routes.expected.yaml",
			runWhen:      func(t *testing.T) { runWhenExpressions(t, ">=3.14.0") },
		},
		{
			name:         "dump skip-defaults: consumers, consumer-groups, consumer-group scoped plugins 3.4",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.4/consumer-group-scoped.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.4/consumer-group-scoped.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.4.0 <3.5.0") },
		},
		{
			name:         "dump skip-defaults: consumers, consumer-groups, consumer-group scoped plugins 3.10+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.10+/consumer-group-scoped.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.10+/consumer-group-scoped.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.10.0 <3.14.0") },
		},
		{
			name:         "dump skip-defaults: consumers, consumer-groups, consumer-group scoped plugins 3.14+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.14+/consumer-group-scoped.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.14+/consumer-group-scoped.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.14.0") },
		},
		{
			name:         "dump skip-defaults: plugins, partials (rla) 3.10+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.10+/plugin-partial.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.10+/plugin-partial.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.10.0 <3.14.0") },
		},
		{
			name:         "dump skip-defaults: plugins, partials (rla) 3.14+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.14+/plugin-partial.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.14+/plugin-partial.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.14.0") },
		},
		{
			name:         "dump skip-defaults: plugins (openid-connect) 3.4+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.4/plugin.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.4/plugin.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.4.0 <3.5.0") },
		},
		{
			name:         "dump skip-defaults: plugins, partials (openid-connect) 3.10+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.10+/plugin-partial-2.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.10+/plugin-partial-2.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.10.0 <3.14.0") },
		},
		{
			name:         "dump skip-defaults: plugins, partials (openid-connect) 3.14+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.14+/plugin-partial-2.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.14+/plugin-partial-2.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.14.0") },
		},
		{
			name:         "dump skip-defaults: plugin rate-limiting 3.10+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.10+/plugin.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.10+/plugin.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.10.0 <3.14.0") },
		},
		{
			name:         "dump skip-defaults: plugin rate-limiting 3.14+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.14+/plugin.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.14+/plugin.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.14.0") },
		},
		{
			name:         "vaults skip-defaults 3.4",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.4/vaults.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.4/vaults.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.4.0 <3.5.0") },
		},
		{
			name:         "dump skip-defaults: non-default shorthand fields",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.10+/plugin-shorthand-non-default.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.10+/plugin-shorthand-non-default.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.10.0 <3.14.0") },
		},
		{
			name:         "dump skip-defaults: non-default shorthand fields 3.14+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.14+/plugin-shorthand-non-default.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.14+/plugin-shorthand-non-default.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.14.0") },
		},
		{
			name:         "vaults skip-defaults 3.10+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.4/vaults.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.4/vaults.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.10.0") },
		},
		{
			name:         "vaults skip-defaults 3.11+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.11+/vaults.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.11+/vaults.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.11.0") },
		},
		{
			name:         "credentials skip-defaults 3.4+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.4/credentials.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.4/credentials.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.4.0 <3.14.0") },
		},
		{
			name:         "credentials skip-defaults 3.14+",
			stateFile:    "testdata/dump/009-skip-defaults/enterprise/3.14+/credentials.yaml",
			expectedFile: "testdata/dump/009-skip-defaults/enterprise/3.14+/credentials.expected.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.14.0") },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.runWhen(t)
			reset(t)
			require.NoError(t, sync(ctx, tc.stateFile))

			dumpArgs := []string{"-o", "-", "--skip-defaults"}
			output, err := dump(dumpArgs...)
			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)

			// ensure that dump can sync back without errors
			require.NoError(t, sync(ctx, tc.expectedFile))

			// dump again
			output, err = dump(dumpArgs...)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

func Test_Dump_Services_TLS_Sans(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		stateFile    string
		expectedFile string
		runWhen      func(t *testing.T)
	}{
		{
			name:         "dump services with TLS SANs",
			stateFile:    "testdata/sync/048-service-tls-sans/kong.yaml",
			expectedFile: "testdata/dump/010-services-tls-sans/kong.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.10.0") },
		},
		{
			name:         "dump services with https but no TLS SANs",
			stateFile:    "testdata/sync/048-service-tls-sans/no-tls-https.yaml",
			expectedFile: "testdata/dump/010-services-tls-sans/no-tls-https.yaml",
			runWhen:      func(t *testing.T) { runWhen(t, "enterprise", ">=3.10.0") },
		},
		{
			name:         "dump services with TLS SANs - konnect",
			stateFile:    "testdata/sync/048-service-tls-sans/kong.yaml",
			expectedFile: "testdata/dump/010-services-tls-sans/konnect.yaml",
			runWhen:      func(t *testing.T) { runWhenKonnect(t) },
		},
		{
			name:         "dump services with https but no TLS SANs - konnect",
			stateFile:    "testdata/sync/048-service-tls-sans/no-tls-https.yaml",
			expectedFile: "testdata/dump/010-services-tls-sans/konnect-no-tls-https.yaml",
			runWhen:      func(t *testing.T) { runWhenKonnect(t) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.runWhen(t)
			reset(t)
			require.NoError(t, sync(ctx, tc.stateFile))

			output, err := dump("-o", "-", "--with-id")
			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

func Test_Dump_GraphqlRateLimitingCostDecorations(t *testing.T) {
	runWhenEnterpriseOrKonnect(t, ">=3.0.0")

	isKonnect := os.Getenv("DECK_KONNECT_EMAIL") != "" ||
		os.Getenv("DECK_KONNECT_PASSWORD") != "" ||
		os.Getenv("DECK_KONNECT_TOKEN") != ""

	if isKonnect {
		runDualTestWithSkipDefaults(t, "Test_Dump_GraphqlRateLimitingCostDecorations",
			testDumpGraphqlRateLimitingCostDecorationsImpl)
	} else {
		testDumpGraphqlRateLimitingCostDecorationsImpl(t)
	}
}

// test scope:
//   - enterprise: >=3.0.0
//   - konnect
func testDumpGraphqlRateLimitingCostDecorationsImpl(t *testing.T) {
	setup(t)

	ctx := context.Background()
	tests := []struct {
		name                string
		stateFile           string
		expectedContains    []string
		notExpectedContains []string
	}{
		{
			name:      "dump single decoration with all fields",
			stateFile: "testdata/dump/012-custom-entities-graphql-ratelimiting/kong-all-fields.yaml",
			expectedContains: []string{
				"graphql_ratelimiting_cost_decorations",
				"Query.posts",
				"add_constant",
				"mul_constant",
				"add_arguments",
				"mul_arguments",
				"limit",
				"offset",
				"first",
				"last",
			},
		},
		{
			name:      "dump multiple decorations",
			stateFile: "testdata/dump/012-custom-entities-graphql-ratelimiting/kong.yaml",
			expectedContains: []string{
				"graphql_ratelimiting_cost_decorations",
				"Query.users",
				"Query.posts",
			},
		},
		{
			name:      "dump empty when none exist",
			stateFile: "testdata/dump/012-custom-entities-graphql-ratelimiting/kong-no-custom-entities.yaml",
			notExpectedContains: []string{
				"graphql_ratelimiting_cost_decorations",
				"custom_entities",
			},
		},
		{
			name:      "dump mixed with other custom entity types",
			stateFile: "testdata/dump/012-custom-entities-graphql-ratelimiting/kong-mixed.yaml",
			expectedContains: []string{
				"graphql_ratelimiting_cost_decorations",
				"Query.users",
				"degraphql_routes",
				"/foo",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				reset(t)
			})
			err := sync(ctx, tc.stateFile)
			require.NoError(t, err)

			output, err := dump("-o", "-")
			require.NoError(t, err)

			for _, expected := range tc.expectedContains {
				assert.Contains(t, output, expected,
					"dump output should contain %q", expected)
			}
			for _, notExpected := range tc.notExpectedContains {
				assert.NotContains(t, output, notExpected,
					"dump output should not contain %q", notExpected)
			}
		})
	}
}

func Test_Dump_KonnectWorkspace(t *testing.T) {
	runWhenKonnect(t)
	setup(t)

	ctx := context.Background()

	tests := []struct {
		name         string
		stateFile    string
		expectedFile string
		dumpFlags    []string
		notContains  string // if set, assert output does NOT contain this string (for isolation tests)
	}{
		{
			name:         "dump empty workspace",
			stateFile:    "testdata/dump/011-konnect-workspace/empty-workspace.yaml",
			expectedFile: "testdata/dump/011-konnect-workspace/expected-empty.yaml",
			dumpFlags:    []string{"-o", "-", "--workspace", "test-workspace"},
		},
		{
			name:         "dump workspace with single entity",
			stateFile:    "testdata/dump/011-konnect-workspace/single-entity.yaml",
			expectedFile: "testdata/dump/011-konnect-workspace/expected-single-entity.yaml",
			dumpFlags:    []string{"-o", "-", "--workspace", "test-workspace"},
		},
		{
			name:         "dump without _workspace field in state file",
			stateFile:    "testdata/dump/011-konnect-workspace/no-workspace.yaml",
			expectedFile: "testdata/dump/011-konnect-workspace/expected-no-workspace.yaml",
			dumpFlags:    []string{"-o", "-"},
		},
		{
			name:         "dump with default workspace name",
			stateFile:    "testdata/dump/011-konnect-workspace/default-workspace.yaml",
			expectedFile: "testdata/dump/011-konnect-workspace/expected-default-workspace.yaml",
			dumpFlags:    []string{"-o", "-"},
		},
		{
			name:        "workspace isolation - entities synced to workspace should not appear in CP-level dump",
			stateFile:   "testdata/dump/011-konnect-workspace/single-entity.yaml",
			dumpFlags:   []string{"-o", "-"},
			notContains: "route-dump-1",
		},
		{
			name:         "dump workspace with multiple entities",
			stateFile:    "testdata/dump/011-konnect-workspace/multiple-entities.yaml",
			expectedFile: "testdata/dump/011-konnect-workspace/expected-multiple-entities.yaml",
			dumpFlags:    []string{"-o", "-", "--workspace", "test-workspace"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			require.NoError(t, sync(ctx, tc.stateFile))

			output, err := dump(tc.dumpFlags...)
			require.NoError(t, err)

			if tc.notContains != "" {
				assert.NotContains(t, output, tc.notContains,
					"Workspace isolation failed: entity should NOT appear in dump")
				return
			}

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}

func Test_Dump_KonnectWorkspace_AllWorkspaces(t *testing.T) {
	runWhenKonnect(t)
	setup(t)

	ctx := context.Background()

	// Create a temp directory for dump output
	tmpDir, err := os.MkdirTemp("", "deck-dump-all-workspaces-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to temp directory for dump output
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Reset and sync entities to two different workspaces
	reset(t)
	require.NoError(t, sync(ctx, originalDir+"/testdata/dump/011-konnect-workspace/workspace1-entity.yaml"))
	require.NoError(t, sync(ctx, originalDir+"/testdata/dump/011-konnect-workspace/workspace2-entity.yaml"))

	// Dump all workspaces
	_, err = dump("--all-workspaces")
	require.NoError(t, err)

	// Verify workspace1.yaml was created with correct content
	workspace1Output, err := os.ReadFile("workspace1.yaml")
	require.NoError(t, err)
	workspace1Expected, err := readFile(originalDir + "/testdata/dump/011-konnect-workspace/expected-workspace1.yaml")
	require.NoError(t, err)
	assert.Equal(t, workspace1Expected, string(workspace1Output))

	// Verify workspace2.yaml was created with correct content
	workspace2Output, err := os.ReadFile("workspace2.yaml")
	require.NoError(t, err)
	workspace2Expected, err := readFile(originalDir + "/testdata/dump/011-konnect-workspace/expected-workspace2.yaml")
	require.NoError(t, err)
	assert.Equal(t, workspace2Expected, string(workspace2Output))

	// Verify default.yaml was created (default workspace should also be dumped)
	defaultOutput, err := os.ReadFile("default.yaml")
	require.NoError(t, err)
	// Default workspace dump should contain _workspace: default
	assert.Contains(t, string(defaultOutput), "_workspace: default",
		"default.yaml should contain _workspace: default")
}

func Test_Dump_Plugin_Conditional(t *testing.T) {
	runWhen(t, "enterprise", ">=3.14.0")
	setup(t)

	ctx := context.Background()
	kongFile := "testdata/sync/003-create-a-plugin/kong-conditional.yaml"
	require.NoError(t, sync(ctx, kongFile))

	output, err := dump("-o", "-", "--with-id")
	require.NoError(t, err)

	expectedFile := "testdata/dump/012-plugin-conditional/expected.yaml"
	expected, err := readFile(expectedFile)
	require.NoError(t, err)
	assert.Equal(t, expected, output)
}

func Test_Dump_Plugin_Conditional_Konnect(t *testing.T) {
	runWhenKonnect(t)
	setup(t)

	ctx := context.Background()
	kongFile := "testdata/sync/003-create-a-plugin/kong-conditional.yaml"
	require.NoError(t, sync(ctx, kongFile))

	output, err := dump("-o", "-", "--with-id")
	require.NoError(t, err)

	expectedFile := "testdata/dump/012-plugin-conditional/expected-konnect.yaml"
	expected, err := readFile(expectedFile)
	require.NoError(t, err)
	assert.Equal(t, expected, output)
}
