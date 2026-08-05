//go:build performance

package performance

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func countHTTPMethods(log string) map[string]int {
	methodCounts := make(map[string]int)

	// Match HTTP request lines like: GET /path HTTP/1.1
	re := regexp.MustCompile(`(?m)^(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD)\s+\/.*\s+HTTP\/[0-9.]+`)

	lines := strings.Split(log, "\n")
	for _, line := range lines {
		if re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				method := matches[1]
				methodCounts[method]++
				methodCounts["total"]++
			}
		}
	}

	return methodCounts
}

// scope
//   - konnect
//   - enterprise
func Test_Sync_Network_Throughput(t *testing.T) {
	tests := []struct {
		name           string
		stateFile      string
		thresholdPOST  int
		thresholdPUT   int
		thresholdTotal int
	}{
		{
			name: "Entities with UUIDs",
			// This file contains 100 services, 100 routes, 10 consumer groups, and 100 consumers in total.
			// Note that real world latency for http response will be about 10x of local instance (which is used in testing)
			// so keeping the acceptable duration low.
			stateFile:      "testdata/sync/regression-entities-with-id.yaml",
			thresholdPUT:   372, // 20% more than 1 request each per entity - we use PUT for create since ID is given.
			thresholdPOST:  120, //  20% more than required - for adding consumers to groups
			thresholdTotal: 525, // Sum of last two + count of GET expected - (310+100+27)*1.2
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setup(t)

			// overwrite default standard output
			rescueStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			var buf bytes.Buffer
			done := make(chan struct{})

			go func() {
				_, _ = io.Copy(&buf, r)
				close(done)
			}()

			err := sync(context.Background(), tc.stateFile, "--verbose", "2")
			require.NoError(t, err)

			w.Close()

			os.Stderr = rescueStderr
			<-done

			result := countHTTPMethods(buf.String())

			fmt.Println(result)

			if result["total"] > tc.thresholdTotal {
				t.Fatalf("expected < %d HTTP requests, sent %d", tc.thresholdTotal, result["total"])
			}
		})
	}
}
