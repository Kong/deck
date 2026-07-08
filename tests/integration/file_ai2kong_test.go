//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func Test_FileAI2Kong(t *testing.T) {
	skipWhenKonnect(t)
	tests := []struct {
		name               string
		inputFile          string
		expectedOutputFile string
	}{
		{
			name:               "convert AI gateway config to Kong gateway config",
			inputFile:          "testdata/file_ai2kong/model-input.yaml",
			expectedOutputFile: "testdata/file_ai2kong/model-output.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := fileAI2Kong("-s", tc.inputFile)

			require.NoError(t, err)

			var expectedOutput interface{}
			var currentOutput interface{}

			content, err := os.ReadFile(tc.expectedOutputFile)
			require.NoError(t, err)

			err = yaml.Unmarshal(content, &expectedOutput)
			require.NoError(t, err)

			err = yaml.Unmarshal([]byte(output), &currentOutput)
			require.NoError(t, err)

			assert.Equal(t, expectedOutput, currentOutput)
		})
	}
}
