//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Dump_SelectTags(t *testing.T) {
	tests := []struct {
		name         string
		stateFile    string
		expectedFile string
	}{
		{
			name:         "dump with select-tags",
			stateFile:    "testdata/dump/001-entities-with-tags/kong.yaml",
			expectedFile: "testdata/dump/001-entities-with-tags/expected.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.0.0")
			teardown := setup(t)
			defer teardown(t)

			assert.NoError(t, sync(tc.stateFile))

			output, err := dump(
				"--select-tag", "managed-by-deck",
				"--select-tag", "org-unit-42",
				"-o", "-",
			)
			assert.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			assert.NoError(t, err)
			assert.Equal(t, output, expected)
		})
	}
}
