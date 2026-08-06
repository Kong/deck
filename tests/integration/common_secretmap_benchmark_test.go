//go:build integration

package integration

import (
	"testing"

	deckFile "github.com/kong/go-database-reconciler/pkg/file"
)

// These benchmarks measure exactly the snippet from cmd/common.go's
// FieldBased dispatch branch:
//
//	mockContent, err := file.GetContentFromFilesWithEnvVars(filenames, file.EnvVarsSkip)
//	secretMap = file.BuildSecretMap(mockContent)
//
// This is the upfront cost FieldBased pays on every sync/diff call that
// ValueBased skips entirely (the guarding `if` is effectively free) — so
// this number IS the difference between the two strategies at this step,
// using the real ~14,000-entity state-{5k,10k,30k}-secrets.yaml files
// through the full file I/O + parse + schema-validate + template pipeline,
// not synthetic in-memory structs.

func BenchmarkCommonGo_BuildSecretMapFromFile_5K(b *testing.B) {
	filenames := []string{getTestStateFile("state-5k-secrets.yaml")}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockContent, err := deckFile.GetContentFromFilesWithEnvVars(filenames, deckFile.EnvVarsSkip)
		if err != nil {
			b.Fatalf("error parsing state file in skipMode for masking: %v", err)
		}
		_ = deckFile.BuildSecretMap(mockContent)
	}
}

func BenchmarkCommonGo_BuildSecretMapFromFile_10K(b *testing.B) {
	filenames := []string{getTestStateFile("state-10k-secrets.yaml")}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockContent, err := deckFile.GetContentFromFilesWithEnvVars(filenames, deckFile.EnvVarsSkip)
		if err != nil {
			b.Fatalf("error parsing state file in skipMode for masking: %v", err)
		}
		_ = deckFile.BuildSecretMap(mockContent)
	}
}

func BenchmarkCommonGo_BuildSecretMapFromFile_30K(b *testing.B) {
	filenames := []string{getTestStateFile("state-30k-secrets.yaml")}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockContent, err := deckFile.GetContentFromFilesWithEnvVars(filenames, deckFile.EnvVarsSkip)
		if err != nil {
			b.Fatalf("error parsing state file in skipMode for masking: %v", err)
		}
		_ = deckFile.BuildSecretMap(mockContent)
	}
}

// BenchmarkCommonGo_FileParseOnly_5K isolates just the file I/O + parse +
// validate + EnvVarsSkip-render cost, without BuildSecretMap — so we can see
// how much of the total is parsing vs. map-building.
func BenchmarkCommonGo_FileParseOnly_5K(b *testing.B) {
	filenames := []string{getTestStateFile("state-5k-secrets.yaml")}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := deckFile.GetContentFromFilesWithEnvVars(filenames, deckFile.EnvVarsSkip)
		if err != nil {
			b.Fatalf("error parsing state file in skipMode for masking: %v", err)
		}
	}
}

func BenchmarkCommonGo_FileParseOnly_10K(b *testing.B) {
	filenames := []string{getTestStateFile("state-10k-secrets.yaml")}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := deckFile.GetContentFromFilesWithEnvVars(filenames, deckFile.EnvVarsSkip)
		if err != nil {
			b.Fatalf("error parsing state file in skipMode for masking: %v", err)
		}
	}
}

func BenchmarkCommonGo_FileParseOnly_30K(b *testing.B) {
	filenames := []string{getTestStateFile("state-30k-secrets.yaml")}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := deckFile.GetContentFromFilesWithEnvVars(filenames, deckFile.EnvVarsSkip)
		if err != nil {
			b.Fatalf("error parsing state file in skipMode for masking: %v", err)
		}
	}
}
