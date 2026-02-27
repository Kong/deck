package performance

import (
	"context"
	"testing"

	"github.com/kong/deck/cmd"
	"github.com/stretchr/testify/require"
)

// setup sets deck env variable to prevent analytics in tests and registers reset
// command with t.Cleanup().
//
// NOTE: Can't be called with tests running t.Parallel() because of the usage
// of t.Setenv().
func setup(t *testing.T) {
	// disable analytics for integration tests
	t.Setenv("DECK_ANALYTICS", "off")
	t.Cleanup(func() {
		reset(t)
	})
}

func sync(ctx context.Context, kongFile string, opts ...string) error {
	deckCmd := cmd.NewRootCmd()
	args := []string{"gateway", "sync", kongFile}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	return deckCmd.ExecuteContext(ctx)
}

func reset(t *testing.T, opts ...string) {
	t.Helper()

	deckCmd := cmd.NewRootCmd()
	args := []string{"gateway", "reset", "--force"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	require.NoError(t, deckCmd.Execute(), "failed to reset Kong's state")
}
