//go:build performance

package performance

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/stretchr/testify/assert"
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
			acceptableExecutionDuration: int64(5000), // 5s
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setup(t)

			/*
						// capture command output to be used during tests
				rescueStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				cmdErr := deckCmd.ExecuteContext(context.Background())

				w.Close()
				out, _ := io.ReadAll(r)
				os.Stdout = rescueStdout
			*/

			// overwrite default standard output
			rescueStderr := os.Stdout
			r, w, _ := os.Pipe()
			os.Stderr = w
			err := sync(context.Background(), tc.stateFile, "--verbose", "2")
			require.NoError(t, err)

			w.Close()
			out, _ := io.ReadAll(r)
			os.Stderr = rescueStderr
			outString := stripansi.Strip(string(out))

			assert.Equal(t, outString, "")
		})
	}
}
