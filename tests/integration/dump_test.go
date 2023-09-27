//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

			assert.NoError(t, sync(tc.stateFile))

			output, err := dump(
				"--select-tag", "managed-by-deck",
				"--select-tag", "org-unit-42",
				"-o", "-",
			)
			assert.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			assert.NoError(t, err)
			assert.Equal(t, output, expected)
		})
	}
}

func Test_Dump_SelectTags_3x(t *testing.T) {
	tests := []struct {
		name         string
		stateFile    string
		expectedFile string
	}{
		{
			name:         "dump with select-tags",
			stateFile:    "testdata/dump/001-entities-with-tags/kong.yaml",
			expectedFile: "testdata/dump/001-entities-with-tags/expected.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.1.0")
			setup(t)

			assert.NoError(t, sync(tc.stateFile))

			output, err := dump(
				"--select-tag", "managed-by-deck",
				"--select-tag", "org-unit-42",
				"-o", "-",
			)
			assert.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			assert.NoError(t, err)
			assert.Equal(t, output, expected)
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
			runWhen:       func(t *testing.T) { runWhen(t, "enterprise", ">=3.5.0") },
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.runWhen(t)
			setup(t)

			assert.NoError(t, sync(tc.stateFile))

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
			assert.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			assert.NoError(t, err)
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

			assert.NoError(t, sync(tc.stateFile))

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
			assert.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			assert.NoError(t, err)
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

			assert.NoError(t, sync(tc.stateFile))

			var (
				output string
				err    error
			)
			flags := []string{"-o", "-", "--with-id"}
			flags = append(flags, tc.flags...)
			output, err = dump(flags...)

			assert.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			assert.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}
