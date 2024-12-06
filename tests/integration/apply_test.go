//go:build integration

package integration

import (
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

var (
	svc1_a = []*kong.Service{
		{
			ID:             kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			Name:           kong.String("svc1"),
			ConnectTimeout: kong.Int(60000),
			Host:           kong.String("mockbin.org"),
			Port:           kong.Int(80),
			Protocol:       kong.String("http"),
			ReadTimeout:    kong.Int(60000),
			Retries:        kong.Int(5),
			WriteTimeout:   kong.Int(60000),
			Tags:           nil,
		},
	}
)

func Test_Apply_3x(t *testing.T) {
	// setup stage

	tests := []struct {
		name          string
		firstFile     string
		secondFile    string
		expectedState string
	}{
		{
			name:          "applies multiple of the same entity",
			firstFile:     "testdata/apply/001-same-type/service-01.yaml",
			secondFile:    "testdata/apply/001-same-type/service-02.yaml",
			expectedState: "testdata/apply/001-same-type/expected-state.yaml",
		},
		{
			name:          "applies different entity types",
			firstFile:     "testdata/apply/002-different-types/service-01.yaml",
			secondFile:    "testdata/apply/002-different-types/plugin-01.yaml",
			expectedState: "testdata/apply/002-different-types/expected-state.yaml",
		},
		{
			name:          "accepts foreign keys",
			firstFile:     "testdata/apply/003-foreign-keys-consumers/consumer-01.yaml",
			secondFile:    "testdata/apply/003-foreign-keys-consumers/plugin-01.yaml",
			expectedState: "testdata/apply/003-foreign-keys-consumers/expected-state.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.0.0")
			setup(t)
			apply(tc.firstFile)
			apply(tc.secondFile)

			out, _ := dump()

			expected, err := readFile(tc.expectedState)
			if err != nil {
				t.Fatalf("failed to read expected state: %v", err)
			}

			assert.Equal(t, expected, out)
		})
	}
}
