//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			name:          "accepts consumer foreign keys",
			firstFile:     "testdata/apply/003-foreign-keys-consumers/consumer-01.yaml",
			secondFile:    "testdata/apply/003-foreign-keys-consumers/plugin-01.yaml",
			expectedState: "testdata/apply/003-foreign-keys-consumers/expected-state.yaml",
		},
		//{
		//	name:          "accepts consumer group foreign keys",
		//	firstFile:     "testdata/apply/004-foreign-keys-consumer-groups/consumer-group-01.yaml",
		//	secondFile:    "testdata/apply/004-foreign-keys-consumer-groups/consumer-01.yaml",
		//	expectedState: "testdata/apply/004-foreign-keys-consumer-groups/expected-state.yaml",
		//},
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
