//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	deckDiff "github.com/kong/go-database-reconciler/pkg/diff"
)

// benchmarkTotalEntities matches TotalEntities in
// scripts/benchmark_generate_testdata.go — the state-*-secrets.yaml files
// were generated with this many total entities, and secret indices are
// templated relative to it, not to maxSecrets.
const benchmarkTotalEntities = 14000

// secretAssignmentIndices mirrors createSecretAssignments in
// scripts/benchmark_generate_testdata.go exactly. The generated state files
// template ${{ env "DECK_RATE_LIMIT_i" }} using the entity's own index i,
// spread evenly across [0, benchmarkTotalEntities) — not a sequential
// [0, maxSecrets) range. Env vars must be set for these exact indices or
// the majority of templates resolve to unset variables and file parsing
// fails outright (silently, since sync() errors are never checked here).
func secretAssignmentIndices(total, secretCount int) []int {
	var indices []int
	if secretCount == 0 || total == 0 {
		return indices
	}
	step := float64(total) / float64(secretCount)
	next := step / 2
	for i := 0; i < secretCount && int(next) < total; i++ {
		indices = append(indices, int(next))
		next += step
	}
	return indices
}

// setupBenchmarkEnvironment sets up all required environment variables for benchmarks
func setupBenchmarkEnvironment(maxSecrets int) {
	for _, i := range secretAssignmentIndices(benchmarkTotalEntities, maxSecrets) {
		os.Setenv(fmt.Sprintf("DECK_RATE_LIMIT_%d", i), fmt.Sprintf("secret_value_%d", i))
		os.Setenv(fmt.Sprintf("DECK_RATE_LIMIT_HOUR_%d", i), fmt.Sprintf("hour_secret_%d", i))
	}
	os.Setenv("DECK_ANALYTICS", "off")
}

// updateBenchmarkSecrets modifies some environment variables to simulate secrets being updated
func updateBenchmarkSecrets(maxSecrets int, updatePercent int) {
	indices := secretAssignmentIndices(benchmarkTotalEntities, maxSecrets)
	updateCount := (len(indices) * updatePercent) / 100
	for _, i := range indices[:updateCount] {
		os.Setenv(fmt.Sprintf("DECK_RATE_LIMIT_%d", i), fmt.Sprintf("updated_secret_%d", i))
	}
}

func getTestStateFile(filename string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", "benchmark", filename)
}

// suppressed I/O for benchmarks
func suppressOutput(fn func() error) error {
	oldStdout, oldStderr := os.Stdout, os.Stderr
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer devNull.Close()
	os.Stdout = devNull
	os.Stderr = devNull
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()
	return fn()

}

// BenchmarkE2E_FieldBased_SyncAndDiff_10K benchmarks field-based masking through complete workflow
// Workflow: Sync → Update secrets → Diff (all measured)
func BenchmarkE2E_FieldBased_SyncAndDiff_10K(b *testing.B) {
	b.StopTimer()
	setupBenchmarkEnvironment(10000)
	deckDiff.SetMaskingStrategy(deckDiff.StrategyFieldBased)

	stateFile := getTestStateFile("state-10k-secrets.yaml")
	if _, err := os.Stat(stateFile); err != nil {
		b.Skipf("State file not found: %s", stateFile)
	}

	ctx := context.Background()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resetBenchMark(b)
		b.StartTimer()
		b.ReportAllocs()
		suppressOutput(func() error {
			return sync(ctx, stateFile)
		})
		b.StopTimer()
		// Update 20% of secrets to simulate environment variable changes
		updateBenchmarkSecrets(10000, 20)

		// Diff: Check what changed with updated secrets (MEASURED)
		b.StartTimer()
		suppressOutput(func() error {
			_, err := diff(stateFile)
			return err
		})
	}

	b.StopTimer()
}

// BenchmarkE2E_ValueBased_SyncAndDiff_10K benchmarks value-based masking through complete workflow
func BenchmarkE2E_ValueBased_SyncAndDiff_10K(b *testing.B) {
	b.StopTimer()
	setupBenchmarkEnvironment(10000)
	deckDiff.SetMaskingStrategy(deckDiff.StrategyValueBased)

	stateFile := getTestStateFile("state-10k-secrets.yaml")
	if _, err := os.Stat(stateFile); err != nil {
		b.Skipf("State file not found: %s", stateFile)
	}

	ctx := context.Background()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resetBenchMark(b)
		b.StartTimer()
		suppressOutput(func() error {
			return sync(ctx, stateFile)
		})
		b.StopTimer()
		updateBenchmarkSecrets(10000, 20)

		b.StartTimer()
		suppressOutput(func() error {
			_, err := diff(stateFile)
			return err
		})
	}

	b.StopTimer()
}

// BenchmarkE2E_FieldBased_SyncAndDiff_5K benchmarks field-based masking on 5K secrets
func BenchmarkE2E_FieldBased_SyncAndDiff_5K(b *testing.B) {
	b.StopTimer()
	setupBenchmarkEnvironment(5000)
	deckDiff.SetMaskingStrategy(deckDiff.StrategyFieldBased)

	stateFile := getTestStateFile("state-5k-secrets.yaml")
	if _, err := os.Stat(stateFile); err != nil {
		b.Skipf("State file not found: %s", stateFile)
	}

	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		suppressOutput(func() error {
			resetBenchMark(b)
			return nil
		})

		b.StartTimer()
		suppressOutput(func() error {
			return sync(ctx, stateFile)
		})

		b.StopTimer()
		updateBenchmarkSecrets(5000, 20)

		b.StartTimer()
		suppressOutput(func() error {
			_, err := diff(stateFile)
			return err
		})
	}

	b.StopTimer()
}

// BenchmarkE2E_ValueBased_SyncAndDiff_5K benchmarks value-based masking on 5K secrets
func BenchmarkE2E_ValueBased_SyncAndDiff_5K(b *testing.B) {
	b.StopTimer()
	setupBenchmarkEnvironment(5000)
	deckDiff.SetMaskingStrategy(deckDiff.StrategyValueBased)

	stateFile := getTestStateFile("state-5k-secrets.yaml")
	if _, err := os.Stat(stateFile); err != nil {
		b.Skipf("State file not found: %s", stateFile)
	}

	ctx := context.Background()
	b.StartTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		suppressOutput(func() error {
			return sync(ctx, stateFile)
		})

		updateBenchmarkSecrets(5000, 20)

		////suppressOutput(func() error {
		//	return sync(ctx, stateFile)
		//})

		suppressOutput(func() error {
			_, err := diff(stateFile)
			return err
		})
	}

	b.StopTimer()
}

// BenchmarkE2E_FieldBased_SyncAndDiff_30K benchmarks field-based masking on 30K secrets
func BenchmarkE2E_FieldBased_SyncAndDiff_30K(b *testing.B) {
	b.StopTimer()
	setupBenchmarkEnvironment(30000)
	deckDiff.SetMaskingStrategy(deckDiff.StrategyFieldBased)

	stateFile := getTestStateFile("state-30k-secrets.yaml")
	if _, err := os.Stat(stateFile); err != nil {
		b.Skipf("State file not found: %s", stateFile)
	}

	ctx := context.Background()
	b.StartTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		suppressOutput(func() error {
			return sync(ctx, stateFile)
		})

		updateBenchmarkSecrets(30000, 20)

		suppressOutput(func() error {
			_, err := diff(stateFile)
			return err
		})
	}

	b.StopTimer()
}

// BenchmarkE2E_ValueBased_SyncAndDiff_30K benchmarks value-based masking on 30K secrets
func BenchmarkE2E_ValueBased_SyncAndDiff_30K(b *testing.B) {
	b.StopTimer()
	setupBenchmarkEnvironment(30000)
	deckDiff.SetMaskingStrategy(deckDiff.StrategyValueBased)

	stateFile := getTestStateFile("state-30k-secrets.yaml")
	if _, err := os.Stat(stateFile); err != nil {
		b.Skipf("State file not found: %s", stateFile)
	}

	ctx := context.Background()
	b.StartTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		suppressOutput(func() error {
			return sync(ctx, stateFile)
		})

		updateBenchmarkSecrets(30000, 20)

		suppressOutput(func() error {
			_, err := diff(stateFile)
			return err
		})
	}

	b.StopTimer()
}
