//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// test scope:
//
//   - AI Gateway >=2.0.0
//
// Test_Dump_AIModels verifies `deck gateway dump` round-trips the ai_model
// entity: after seeding Kong with ai_models, the dumped file must contain the
// same entities. ai_model is only available on AI Gateway instances.
func Test_Dump_AIModels(t *testing.T) {
	runWhenAIGateway(t, ">=2.0.0")
	setup(t)

	ctx := context.Background()

	// ID and CreatedAt/UpdatedAt are server-assigned, so they are ignored.
	ignoreFields := []cmp.Option{
		cmpopts.IgnoreFields(kong.AIModel{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.SortSlices(func(a, b file.FAIModel) bool { return *a.Name < *b.Name }),
		cmpopts.SortSlices(func(a, b *string) bool { return *a < *b }),
		cmpopts.EquateEmpty(),
	}

	expected := []file.FAIModel{
		{
			AIModel: kong.AIModel{
				Name:  kong.String("gpt-5"),
				Alias: kong.String("@openai/gpt-5"),
			},
		},
		{
			AIModel: kong.AIModel{
				Name:  kong.String("claude-opus"),
				Alias: kong.String("@anthropic/claude-opus"),
				Tags:  kong.StringSlice("ai", "anthropic"),
			},
		},
	}

	reset(t)
	require.NoError(t, sync(ctx, "testdata/sync/056-ai-models/kong.yaml"))

	output, err := dump("-o", "-")
	require.NoError(t, err)

	content := parseAIState(t, output)
	if diff := cmp.Diff(content.AIModels, expected, ignoreFields...); diff != "" {
		t.Errorf("unexpected ai_models dump diff:\n%s", diff)
	}
}

// Test_AIDump exercises `deck ai dump`, which reads the AI-managed entities
// (tagged 'managed-by:deck-ai') from Kong and reverts them back to AI Gateway
// format.
//
// `deck ai dump` does not reproduce the original source file byte-for-byte (it
// fills defaults, renames providers, and drops presentation-only fields), so
// there is no static fixture to compare against. Instead we assert the property
// that matters: `ai dump` is a faithful inverse of `ai sync`. After seeding Kong
// from an AI Gateway source, the dumped configuration must, when synced again,
// reproduce exactly the same AI-managed Kong state. If `ai dump` dropped or
// mangled any entity, the round-tripped state would diverge.
//
// State is compared via `deck gateway dump` (scoped to the managed-by:deck-ai
// tag, IDs stripped) using the same structural comparison as Test_AISync.
func Test_AIDump(t *testing.T) {
	runWhenAIGateway(t, ">=2.0.0")
	setup(t)

	ctx := context.Background()

	tests := []struct {
		name      string
		inputFile string
	}{
		{
			name:      "models",
			inputFile: "testdata/file_ai2kong/01-models/input.yaml",
		},
		{
			name:      "identity and policies",
			inputFile: "testdata/file_ai2kong/02-identity-and-policies/input.yaml",
		},
		{
			name:      "agents",
			inputFile: "testdata/file_ai2kong/03-agents/input.yaml",
		},
		{
			name:      "mcp",
			inputFile: "testdata/file_ai2kong/04-mcp/input.yaml",
		},
		{
			name:      "identity providers",
			inputFile: "testdata/file_ai2kong/05-identity-providers/input.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Seed Kong from the AI Gateway source and capture the resulting
			// AI-managed state as the reference.
			reset(t)
			require.NoError(t, aiSync(ctx, tc.inputFile))
			reference, err := dump("--select-tag", managedByAIDeckTag, "-o", "-")
			require.NoError(t, err)

			// `ai dump` reverts the state back to AI Gateway format.
			aiConfig, err := aiDump("-o", "-")
			require.NoError(t, err)
			require.NotEmpty(t, aiConfig)

			// Dumping the same state again must be deterministic.
			aiConfigAgain, err := aiDump("-o", "-")
			require.NoError(t, err)
			assert.Equal(t, aiConfig, aiConfigAgain)

			// Round-trip: syncing the dumped AI Gateway config into a fresh Kong
			// must reproduce the same AI-managed state.
			roundTripFile := filepath.Join(t.TempDir(), "ai-dump.yaml")
			require.NoError(t, os.WriteFile(roundTripFile, []byte(aiConfig), 0o600))

			reset(t)
			require.NoError(t, aiSync(ctx, roundTripFile))
			roundTripped, err := dump("--select-tag", managedByAIDeckTag, "-o", "-")
			require.NoError(t, err)

			assertAIStateEqual(t, reference, roundTripped)
		})
	}
}
