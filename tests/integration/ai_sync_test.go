//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/require"
)

// test scope:
//
//   - AI Gateway >=2.0.0
//
// ai_model is only available on AI Gateway instances, so this test is skipped
// on regular Kong / Kong Enterprise and on Konnect.
func Test_Sync_AIModels(t *testing.T) {
	runWhenAIGateway(t, ">=2.0.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)
	ctx := context.Background()

	// AIModel CreatedAt/UpdatedAt are server-assigned timestamps.
	ignoreFields := []cmp.Option{
		cmpopts.IgnoreFields(kong.AIModel{}, "CreatedAt", "UpdatedAt"),
	}

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates ai_models",
			kongFile: "testdata/sync/056-ai-models/kong.yaml",
			expectedState: utils.KongRawState{
				AIModels: []*kong.AIModel{
					{
						ID:    kong.String("3c9d1e2f-4a5b-6c7d-8e9f-0a1b2c3d4e5f"),
						Name:  kong.String("claude-opus"),
						Alias: kong.String("@anthropic/claude-opus"),
						Tags:  kong.StringSlice("ai", "anthropic"),
					},
					{
						ID:    kong.String("8b4a7b3e-1b2c-4d5e-9f6a-0c1d2e3f4a5b"),
						Name:  kong.String("gpt-5"),
						Alias: kong.String("@openai/gpt-5"),
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			err := sync(ctx, tc.kongFile)
			require.NoError(t, err)

			testKongState(t, client, false, false, tc.expectedState, ignoreFields)

			// re-sync with no error
			err = sync(ctx, tc.kongFile)
			require.NoError(t, err)
		})
	}
}

// aiSyncTestCase pairs an AI Gateway source (inputFile) with its pre-converted
// Kong configuration (outputFile) for Test_AISync.
type aiSyncTestCase struct {
	name       string
	inputFile  string
	outputFile string
}

// runAISyncCases exercises `deck ai sync`, which converts an AI Gateway state
// file to Kong configuration and syncs it directly to Kong.
//
// The testdata under testdata/file_ai2kong/<case> holds an AI Gateway source
// (input.yaml) alongside its converted Kong configuration (output.yaml). Since
// `ai sync` is the equivalent of `deck file ai2kong` followed by
// `deck gateway sync`, syncing input.yaml via `ai sync` must reach exactly the
// same AI-managed Kong state as syncing output.yaml directly. We capture that
// direct sync as the expected state, assert that `ai sync` converges to it, and
// then assert that re-running `ai sync` keeps the state consistent.
//
// State is compared via `deck gateway dump` (scoped to the managed_by:deck-ai
// tag, IDs stripped) so the comparison is independent of server-assigned IDs.
func runAISyncCases(t *testing.T, tests []aiSyncTestCase) {
	t.Helper()
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Establish the expected AI-managed state by syncing the converted
			// (ai2kong) configuration directly.
			reset(t)
			require.NoError(t, sync(ctx, tc.outputFile))
			expected, err := dump("--select-tag", managedByAIDeckTag, "-o", "-")
			require.NoError(t, err)

			// `ai sync` of the AI Gateway source must reach the same state.
			reset(t)
			require.NoError(t, aiSync(ctx, tc.inputFile))
			afterSync, err := dump("--select-tag", managedByAIDeckTag, "-o", "-")
			require.NoError(t, err)
			assertAIStateEqual(t, expected, afterSync)

			// Re-syncing must succeed and keep the state consistent.
			require.NoError(t, aiSync(ctx, tc.inputFile))
			afterResync, err := dump("--select-tag", managedByAIDeckTag, "-o", "-")
			require.NoError(t, err)
			assertAIStateEqual(t, afterSync, afterResync)
		})
	}
}

func Test_AISync(t *testing.T) {
	runWhenAIGateway(t, ">=2.0.0")
	setup(t)

	runAISyncCases(t, []aiSyncTestCase{
		{
			name:       "models",
			inputFile:  "testdata/file_ai2kong/01-models/input.yaml",
			outputFile: "testdata/file_ai2kong/01-models/output.yaml",
		},
		{
			name:       "identity and policies",
			inputFile:  "testdata/file_ai2kong/02-identity-and-policies/input.yaml",
			outputFile: "testdata/file_ai2kong/02-identity-and-policies/output.yaml",
		},
		{
			name:       "agents",
			inputFile:  "testdata/file_ai2kong/03-agents/input.yaml",
			outputFile: "testdata/file_ai2kong/03-agents/output.yaml",
		},
		{
			name:       "mcp",
			inputFile:  "testdata/file_ai2kong/04-mcp/input.yaml",
			outputFile: "testdata/file_ai2kong/04-mcp/output.yaml",
		},
		{
			name:       "auth strategies",
			inputFile:  "testdata/file_ai2kong/05-auth-strategies/input.yaml",
			outputFile: "testdata/file_ai2kong/05-auth-strategies/output.yaml",
		},
		{
			name:       "certificate and SNI",
			inputFile:  "testdata/file_ai2kong/06-certificate-and-sni/input.yaml",
			outputFile: "testdata/file_ai2kong/06-certificate-and-sni/output.yaml",
		},
		{
			name:       "ca certificates",
			inputFile:  "testdata/file_ai2kong/07-ca-certificates/input.yaml",
			outputFile: "testdata/file_ai2kong/07-ca-certificates/output.yaml",
		},
	})
}

// Test_AISync_AIGateway21 covers the AI Gateway 2.1 fields that have an AI
// Gateway entity-model representation. It is split out from Test_AISync so the
// 2.0.x image, whose plugin schemas reject those fields, skips it.
func Test_AISync_AIGateway21(t *testing.T) {
	runWhenAIGateway(t, ">=2.1.0")
	setup(t)

	runAISyncCases(t, []aiSyncTestCase{
		{
			name:       "model cost lists",
			inputFile:  "testdata/file_ai2kong/08-model-cost-lists/input.yaml",
			outputFile: "testdata/file_ai2kong/08-model-cost-lists/output.yaml",
		},
		{
			name:       "mcp protocol 2.1 fields",
			inputFile:  "testdata/file_ai2kong/09-mcp-protocol-2-1-fields/input.yaml",
			outputFile: "testdata/file_ai2kong/09-mcp-protocol-2-1-fields/output.yaml",
		},
		{
			name:       "auth strategy bearer header",
			inputFile:  "testdata/file_ai2kong/10-auth-strategy-bearer-header/input.yaml",
			outputFile: "testdata/file_ai2kong/10-auth-strategy-bearer-header/output.yaml",
		},
		{
			name:       "policy condition",
			inputFile:  "testdata/file_ai2kong/11-policy-condition/input.yaml",
			outputFile: "testdata/file_ai2kong/11-policy-condition/output.yaml",
		},
	})
}

// Test_AISync_MultipleFiles exercises `deck ai sync` with more than one source
// file, including a mix of formats (one YAML and one JSON).
func Test_AISync_MultipleFiles(t *testing.T) {
	runWhenAIGateway(t, ">=2.0.0")
	setup(t)

	ctx := context.Background()

	// One YAML source (an agent) and one JSON source (an MCP server); each
	// converts to distinctly-named Kong entities so the merge is conflict-free.
	inputFiles := []string{
		"testdata/ai_sync/01-multiple-files/input1.yaml",
		"testdata/ai_sync/01-multiple-files/input2.json",
	}
	outputFiles := []string{
		"testdata/ai_sync/01-multiple-files/output1.yaml",
		"testdata/ai_sync/01-multiple-files/output2.yaml",
	}

	// Establish the expected AI-managed state by syncing the converted
	// (ai2kong) configurations directly.
	reset(t)
	require.NoError(t, sync(ctx, outputFiles[0], outputFiles[1:]...))
	expected, err := dump("--select-tag", managedByAIDeckTag, "-o", "-")
	require.NoError(t, err)

	// `ai sync` of the two AI Gateway sources must reach the same state.
	reset(t)
	require.NoError(t, aiSync(ctx, inputFiles...))
	afterSync, err := dump("--select-tag", managedByAIDeckTag, "-o", "-")
	require.NoError(t, err)
	assertAIStateEqual(t, expected, afterSync)

	// Re-syncing must succeed and keep the state consistent.
	require.NoError(t, aiSync(ctx, inputFiles...))
	afterResync, err := dump("--select-tag", managedByAIDeckTag, "-o", "-")
	require.NoError(t, err)
	assertAIStateEqual(t, afterSync, afterResync)
}
