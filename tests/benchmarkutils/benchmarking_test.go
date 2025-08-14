//go:build benchmark

package benchmarkutils

import (
	"strings"
	"sync/atomic"
	"testing"
)

func runBenchmark(b *testing.B, name string, fn func(b *testing.B) error) {
	var errCount int64

	b.Run(name, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := fn(b)
			if err != nil {
				if !strings.Contains(err.Error(), "EOF") {
					atomic.AddInt64(&errCount, 1)
				}
			}
		}

		total := b.N
		errors := atomic.LoadInt64(&errCount)

		b.ReportMetric(float64(errors)/float64(total), "errors/op")
	})
}

func BenchmarkSync(b *testing.B) {
	runBenchmark(b, "sync", func(b *testing.B) error {
		setup(b)

		b.StartTimer()
		err := sync(b.Context(), "testdata/sync/kong.yaml")
		if err != nil {
			return err
		}
		b.StopTimer()

		return nil
	})
}
