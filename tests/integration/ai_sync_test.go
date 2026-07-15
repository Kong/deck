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

// Test_AISync exercises `deck ai sync`, which converts an AI Gateway state file
// to Kong configuration and syncs it directly to Kong.
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
func Test_AISync(t *testing.T) {
	runWhenAIGateway(t, ">=2.0.0")
	setup(t)

	ctx := context.Background()

	tests := []struct {
		name       string
		inputFile  string
		outputFile string
	}{
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
			name:       "identity providers",
			inputFile:  "testdata/file_ai2kong/05-identity-providers/input.yaml",
			outputFile: "testdata/file_ai2kong/05-identity-providers/output.yaml",
		},
	}
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
