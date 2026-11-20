//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const bulkTestdata = "testdata/validate/004-konnect-bulk-config/"

// validateBulk fails if the control plane does not have the config validation endpoint.
// Without --require-bulk-validation, a missing endpoint makes these tests pass with the per-entity validation.
func validateBulk(opts ...string) error {
	return validate(ONLINE, append(opts, "--require-bulk-validation")...)
}

func Test_Validate_Konnect_BulkConfig(t *testing.T) {
	runWhenKonnect(t)
	setDefaultKonnectControlPlane(t)
	setup(t)

	tests := []struct {
		name           string
		stateFile      string
		additionalArgs []string
		errorExpected  bool
		errorStrings   []string
		notInError     []string
	}{
		{
			name:      "valid configuration with each supported entity type",
			stateFile: bulkTestdata + "valid.yaml",
		},
		{
			name:           "valid configuration in a workspace",
			stateFile:      bulkTestdata + "valid.yaml",
			additionalArgs: []string{"--workspace=default"},
		},
		{
			name:           "plugins only, with the entities they refer to",
			stateFile:      bulkTestdata + "valid.yaml",
			additionalArgs: []string{"--online-entities-list=Plugins"},
		},
		{
			name:           "entity list without the conflicting plugins",
			stateFile:      bulkTestdata + "duplicate-instance-name.yaml",
			additionalArgs: []string{"--online-entities-list=Services"},
		},
		{
			name:          "plugins with the same instance_name",
			stateFile:     bulkTestdata + "duplicate-instance-name.yaml",
			errorExpected: true,
			errorStrings: []string{
				"conflicting entity found",
				"0b5c1d3e-0003-4000-8000-000000000101",
				"0b5c1d3e-0003-4000-8000-000000000102",
			},
		},
		{
			name:          "plugin config with an unknown field",
			stateFile:     bulkTestdata + "plugin-unknown-field.yaml",
			errorExpected: true,
			errorStrings:  []string{"validate entity 'plugin (", "bulk_unknown_field"},
		},
		{
			name:          "plugin config with a field of the wrong type",
			stateFile:     bulkTestdata + "plugin-wrong-type.yaml",
			errorExpected: true,
			errorStrings:  []string{"validate entity 'plugin (", "minute"},
		},
		{
			name:          "plugin with an instance_name that is not valid",
			stateFile:     bulkTestdata + "plugin-invalid-instance-name.yaml",
			errorExpected: true,
			errorStrings:  []string{"validate entity 'plugin (", "instance_name"},
		},
		{
			name:          "each invalid entity is reported",
			stateFile:     bulkTestdata + "multiple-errors.yaml",
			errorExpected: true,
			errorStrings: []string{
				"bulk-bad-port", "port",
				"bulk-bad-retries", "retries",
				"0b5c1d3e-000c-4000-8000-000000000701", "ttl",
			},
			notInError: []string{"bulk-secret-key-value"},
		},
		{
			name:      "valid cloned plugins",
			stateFile: "testdata/sync/054-cloned-plugin-definitions/kong.yaml",
		},
		{
			name:          "cloned plugin config is validated with the schema of its ref",
			stateFile:     bulkTestdata + "cloned-plugin-invalid-config.yaml",
			errorExpected: true,
			errorStrings:  []string{"bulk_unknown_field"},
		},
		{
			name:          "cloned plugin with a ref that cannot be cloned",
			stateFile:     bulkTestdata + "cloned-plugin-unsupported-ref.yaml",
			errorExpected: true,
			errorStrings:  []string{"does not support cloning"},
		},
		{
			name:          "plugin of a cloned plugin that is not defined",
			stateFile:     bulkTestdata + "cloned-plugin-undefined.yaml",
			errorExpected: true,
			errorStrings:  []string{"bulk-undefined-clone"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBulk(append([]string{tc.stateFile}, tc.additionalArgs...)...)

			if !tc.errorExpected {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			for _, s := range tc.errorStrings {
				assert.Contains(t, err.Error(), s)
			}
			for _, s := range tc.notInError {
				assert.NotContains(t, err.Error(), s)
			}
		})
	}
}

// The configuration validation endpoint validates the file as a full configuration.
// Entities of the control plane that the file replaces do not conflict with it.
func Test_Validate_Konnect_BulkConfig_IgnoresCurrentEntities(t *testing.T) {
	runWhenKonnect(t)
	setDefaultKonnectControlPlane(t)
	setup(t)
	reset(t)

	require.NoError(t, sync(context.Background(), bulkTestdata+"current-state.yaml"))

	require.NoError(t, validateBulk(bulkTestdata+"reuses-current-names.yaml"))
}

func Test_Validate_Konnect_BulkConfig_Partials(t *testing.T) {
	runWhenKonnect(t)
	setDefaultKonnectControlPlane(t)
	setup(t)
	reset(t)
	require.NoError(t, sync(context.Background(), "testdata/validate/001-partials/partials.yaml"))

	t.Run("plugin with partials uses the per-entity validation", func(t *testing.T) {
		require.NoError(t, validateBulk(bulkTestdata+"partials-valid.yaml"))
	})

	t.Run("conflict between a plugin with partials and a bulk plugin", func(t *testing.T) {
		err := validateBulk(bulkTestdata + "partials-conflict.yaml")

		require.Error(t, err)
		assert.Contains(t, err.Error(),
			`validate entity 'plugins (0b5c1d3e-0003-4000-8000-000000000a01)': `+
				`instance_name "bulk-partial-plugin" is also used by plugins 0b5c1d3e-0003-4000-8000-000000000a02`)
	})
}

func Test_Validate_Konnect_BulkConfig_Chunks(t *testing.T) {
	runWhenKonnect(t)
	setDefaultKonnectControlPlane(t)
	setup(t)

	const trees = 1500

	t.Run("references are in each chunk", func(t *testing.T) {
		stateFile := writeBulkConfig(t, trees, false)

		require.NoError(t, validateBulk(stateFile, "--config-chunk-size=1"))
	})

	t.Run("conflicting entities at both ends of the file are in one chunk", func(t *testing.T) {
		stateFile := writeBulkConfig(t, trees, true)

		err := validateBulk(stateFile, "--config-chunk-size=1")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "conflicting entity found")
		assert.Contains(t, err.Error(), bulkTreeID(3, 0))
		assert.Contains(t, err.Error(), bulkTreeID(3, trees-1))
	})

	t.Run("chunk size less than 1", func(t *testing.T) {
		err := validateBulk(bulkTestdata+"valid.yaml", "--config-chunk-size=0")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--config-chunk-size cannot be less than 1")
	})
}

func Test_Validate_Konnect_BulkConfig_HiddenErrors(t *testing.T) {
	runWhenKonnect(t)
	setDefaultKonnectControlPlane(t)
	setup(t)

	services := make([]map[string]any, 0, 20)
	for i := range 20 {
		services = append(services, map[string]any{
			"id":   bulkTreeID(1, i),
			"name": fmt.Sprintf("bulk-bad-port-%d", i),
			"host": "example.com",
			"port": 70000,
		})
	}
	stateFile := writeStateFile(t, map[string]any{"_format_version": "3.0", "services": services})

	err := validateBulk(stateFile)

	require.Error(t, err)
	assert.Contains(t, err.Error(),
		"validate configuration: 4 more errors are not shown. Fix the errors above and validate again")
}

func Test_Validate_RequireBulkValidation_NotKonnect(t *testing.T) {
	setup(t)
	runWhen(t, "kong", ">=2.8.0")

	err := validateBulk("testdata/validate/kong3x.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--require-bulk-validation works only with Konnect")
}

// bulkTreeID returns the ID of entity kind of tree i.
func bulkTreeID(kind, i int) string {
	return fmt.Sprintf("0b5c1d3e-%04x-4000-8000-%012d", 0x100+kind, i)
}

// writeBulkConfig writes trees services, each with a route and a plugin.
// The tags make the file larger than 2 MB.
// If conflict is true, the plugins of the first and the last tree have the same instance_name.
func writeBulkConfig(t *testing.T, trees int, conflict bool) string {
	t.Helper()
	tags := func(i int) []string {
		out := make([]string, 0, 5)
		for j := range 5 {
			out = append(out, fmt.Sprintf("bulk-%d-%d-%s", i, j, strings.Repeat("x", 80)))
		}
		return out
	}
	services := make([]map[string]any, 0, trees)
	for i := range trees {
		instanceName := fmt.Sprintf("bulk-rl-%d", i)
		if conflict && i == trees-1 {
			instanceName = "bulk-rl-0"
		}
		services = append(services, map[string]any{
			"id":   bulkTreeID(1, i),
			"name": fmt.Sprintf("bulk-chunk-svc-%d", i),
			"host": "example.com",
			"tags": tags(i),
			"routes": []map[string]any{{
				"id":    bulkTreeID(2, i),
				"name":  fmt.Sprintf("bulk-chunk-route-%d", i),
				"paths": []string{fmt.Sprintf("/bulk-chunk-%d", i)},
				"tags":  tags(i),
			}},
			"plugins": []map[string]any{{
				"id":            bulkTreeID(3, i),
				"name":          "rate-limiting",
				"instance_name": instanceName,
				"config":        map[string]any{"minute": 5, "policy": "local"},
				"tags":          tags(i),
			}},
		})
	}
	stateFile := writeStateFile(t, map[string]any{"_format_version": "3.0", "services": services})
	info, err := os.Stat(stateFile)
	require.NoError(t, err)
	require.Greater(t, info.Size(), int64(2<<20), "the file must need more than one chunk of 1 MB")
	return stateFile
}

func writeStateFile(t *testing.T, content map[string]any) string {
	t.Helper()
	b, err := json.Marshal(content)
	require.NoError(t, err)
	stateFile := filepath.Join(t.TempDir(), "kong.json")
	require.NoError(t, os.WriteFile(stateFile, b, 0o600))
	return stateFile
}
