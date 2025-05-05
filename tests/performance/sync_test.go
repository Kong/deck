//go:build performance

package performance

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Sync_Execution_Duration_Simple(t *testing.T) {
	tests := []struct {
		name                        string
		stateFile                   string
		acceptableExecutionDuration int64
	}{
		{
			name: "Entities with UUIDs",
			// This file contains 100 services, 10 consumer groups, and 100 consumers in total.
			// Note that real world latency for http response will be about 10x of local instance (which is used in testing)
			// so keeping the acceptable duration low.
			stateFile:                   "testdata/sync/regression-entities-with-id.yaml",
			acceptableExecutionDuration: int64(1000), // 1s
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setup(t)
			start := time.Now()
			err := sync(context.Background(), tc.stateFile)
			elapsed := time.Since(start)
			require.NoError(t, err)
			if elapsed.Milliseconds() > tc.acceptableExecutionDuration {
				t.Errorf("expected execution time for sync to be < %d ms", tc.acceptableExecutionDuration)
			}
		})
	}
}
