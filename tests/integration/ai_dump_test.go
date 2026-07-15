//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
