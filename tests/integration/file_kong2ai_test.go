//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

// fileKong2AITestCase points at a fixture directory shared with Test_FileAI2Kong:
// output.yaml is the Kong decK file fed to `file kong2ai`, and input.yaml is the
// AI Gateway source it was originally generated from.
type fileKong2AITestCase struct {
	name string
	dir  string
}

// The generated AI Gateway file must be accepted by `ai sync`, and must
// yield the same AI-managed Kong state as the original AI Gateway source.
func runFileKong2AICases(t *testing.T, tests []fileKong2AITestCase) {
	t.Helper()
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			require.NoError(t, aiSync(ctx, filepath.Join(tc.dir, "input.yaml")))
			reference, err := dump("--select-tag", managedByAIDeckTag, "-o", "-")
			require.NoError(t, err)

			aiConfig, err := fileKong2AI("-s", filepath.Join(tc.dir, "output.yaml"))
			require.NoError(t, err)
			require.NotEmpty(t, aiConfig)

			migratedFile := filepath.Join(t.TempDir(), "kong2ai.yaml")
			require.NoError(t, os.WriteFile(migratedFile, []byte(aiConfig), 0o600))

			reset(t)
			require.NoError(t, aiSync(ctx, migratedFile))

			migrated, err := dump("--select-tag", managedByAIDeckTag, "-o", "-")
			require.NoError(t, err)

			assertAIStateEqual(t, reference, migrated)
		})
	}
}

func Test_FileKong2AI(t *testing.T) {
	runWhenAIGateway(t, ">=2.0.0")
	setup(t)

	runFileKong2AICases(t, []fileKong2AITestCase{
		{
			name: "models",
			dir:  "testdata/file_ai2kong/01-models",
		},
		{
			name: "identity and policies",
			dir:  "testdata/file_ai2kong/02-identity-and-policies",
		},
		{
			name: "agents",
			dir:  "testdata/file_ai2kong/03-agents",
		},
		{
			name: "mcp",
			dir:  "testdata/file_ai2kong/04-mcp",
		},
		{
			name: "auth strategies",
			dir:  "testdata/file_ai2kong/05-auth-strategies",
		},
		{
			name: "certificate and SNI",
			dir:  "testdata/file_ai2kong/06-certificate-and-sni",
		},
		{
			name: "ca certificates",
			dir:  "testdata/file_ai2kong/07-ca-certificates",
		},
	})
}

func Test_FileKong2AI_AIGateway21(t *testing.T) {
	runWhenAIGateway(t, ">=2.1.0")
	setup(t)

	runFileKong2AICases(t, []fileKong2AITestCase{
		{
			name: "model cost lists",
			dir:  "testdata/file_ai2kong/08-model-cost-lists",
		},
		{
			name: "mcp protocol 2.1 fields",
			dir:  "testdata/file_ai2kong/09-mcp-protocol-2-1-fields",
		},
		{
			name: "auth strategy bearer header",
			dir:  "testdata/file_ai2kong/10-auth-strategy-bearer-header",
		},
		{
			name: "policy condition",
			dir:  "testdata/file_ai2kong/11-policy-condition",
		},
	})
}

func Test_FileKong2AI_AIGateway22(t *testing.T) {
	runWhenAIGateway(t, ">=2.2.0")
	setup(t)

	runFileKong2AICases(t, []fileKong2AITestCase{
		{
			name: "skills api",
			dir:  "testdata/file_ai2kong/12-skills-api",
		},
		{
			name: "passthrough format",
			dir:  "testdata/file_ai2kong/13-passthrough-format",
		},
		{
			name: "typesafe provider decisions",
			dir:  "testdata/file_ai2kong/14-typesafe-decisions",
		},
		{
			name: "typesafe provider decisions with multiple aliases",
			dir:  "testdata/file_ai2kong/15-typesafe-decisions-multi-alias",
		},
	})
}

// Custom policy schema needs to be registered before use. So, we will test it differently.
func Test_FileKong2AI_CustomPolicy(t *testing.T) {
	runWhenAIGateway(t, ">=2.2.0")
	setup(t)

	ctx := context.Background()
	const (
		customPolicyFile = "testdata/file_ai2kong/16-custom-policy/custom-policy.yaml"
		policyFile       = "testdata/file_ai2kong/16-custom-policy/policy.yaml"
		kongFile         = "testdata/file_ai2kong/16-custom-policy/output.yaml"
	)

	reset(t)
	require.NoError(t, aiSync(ctx, customPolicyFile, "--include-policy-definitions"))
	time.Sleep(pluginDefinitionSyncDelay)
	require.NoError(t, aiSync(ctx, policyFile, "--include-policy-definitions"))
	reference, err := dump("--select-tag", managedByAIDeckTag, "-o", "-", "--include-plugin-definitions")
	require.NoError(t, err)

	aiConfig, err := fileKong2AI("-s", kongFile)
	require.NoError(t, err)
	require.NotEmpty(t, aiConfig)

	migratedFile := filepath.Join(t.TempDir(), "kong2ai.yaml")
	require.NoError(t, os.WriteFile(migratedFile, []byte(aiConfig), 0o600))

	reset(t)
	require.NoError(t, aiSync(ctx, customPolicyFile, "--include-policy-definitions"))
	time.Sleep(pluginDefinitionSyncDelay)
	require.NoError(t, aiSync(ctx, migratedFile, "--include-policy-definitions"))
	migrated, err := dump("--select-tag", managedByAIDeckTag, "-o", "-", "--include-plugin-definitions")
	require.NoError(t, err)

	assertAIStateEqual(t, reference, migrated)
}

func Test_FileKong2AI_OutputFormats(t *testing.T) {
	runWhenAIGateway(t, ">=2.0.0")
	setup(t)

	const dir = "testdata/file_ai2kong/02-identity-and-policies"
	kongYAMLFile := filepath.Join(dir, "output.yaml")

	t.Run("json source gives the same result as yaml source", func(t *testing.T) {
		kongJSONFile := filepath.Join(t.TempDir(), "kong.json")
		kongJSON, err := yaml.YAMLToJSON([]byte(mustReadFile(t, kongYAMLFile)))
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(kongJSONFile, kongJSON, 0o600))

		fromYAML, err := fileKong2AI("-s", kongYAMLFile)
		require.NoError(t, err)
		fromJSON, err := fileKong2AI("-s", kongJSONFile)
		require.NoError(t, err)

		assert.YAMLEq(t, fromYAML, fromJSON)
	})
}
