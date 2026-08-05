package benchmarkutils

import (
	"context"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/kong/deck/cmd"
)

func setup(b *testing.B) {
	b.Setenv("DECK_ANALYTICS", "off")
	b.Cleanup(func() {
		reset(b)
	})
}

func reset(b *testing.B, opts ...string) {
	b.Helper()

	deckCmd := cmd.NewRootCmd()
	args := []string{"gateway", "reset", "-f"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	err := deckCmd.Execute()
	if err != nil {
		b.Fatalf("failed to reset Kong's state: %s", err.Error())
	}
}

func sync(ctx context.Context, kongFile string, opts ...string) error {
	deckCmd := cmd.NewRootCmd()
	args := []string{"gateway", "sync", kongFile}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	_, w, _ := os.Pipe()
	color.Output = w

	cmdErr := deckCmd.ExecuteContext(ctx)

	w.Close()

	return cmdErr
}
