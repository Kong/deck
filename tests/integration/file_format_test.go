package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func Test_FileFormat_DBlessToDeck(t *testing.T) {
	tests := []struct {
		name               string
		inputFile          string
		expectedOutputFile string
		errorExpected      bool
		errorString        string
	}{
		{
			name:               "converts DBless consumer groups to decK format",
			inputFile:          "testdata/file-format/dbless-input.yaml",
			expectedOutputFile: "testdata/file-format/deck-input.yaml",
		},
		{
			name:               "file with no consumer groups passes through unchanged",
			inputFile:          "testdata/file-format/no-consumer-groups.yaml",
			expectedOutputFile: "testdata/file-format/no-consumer-groups.yaml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := fileFormat("deck", tc.inputFile)

			if tc.errorExpected {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorString)
				return
			}

			require.NoError(t, err)

			content, err := os.ReadFile(tc.expectedOutputFile)
			require.NoError(t, err)

			var expected, actual interface{}
			require.NoError(t, yaml.Unmarshal(content, &expected))
			require.NoError(t, yaml.Unmarshal([]byte(output), &actual))
			assert.Equal(t, expected, actual)
		})
	}
}

func Test_FileFormat_DeckToDBless(t *testing.T) {
	tests := []struct {
		name               string
		inputFile          string
		expectedOutputFile string
		errorExpected      bool
		errorString        string
	}{
		{
			name:               "converts decK consumer groups to DBless format",
			inputFile:          "testdata/file-format/deck-input.yaml",
			expectedOutputFile: "testdata/file-format/dbless-input.yaml",
		},
		{
			name:               "file with no consumer groups passes through unchanged",
			inputFile:          "testdata/file-format/no-consumer-groups.yaml",
			expectedOutputFile: "testdata/file-format/no-consumer-groups.yaml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := fileFormat("dbless", tc.inputFile)

			if tc.errorExpected {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorString)
				return
			}

			require.NoError(t, err)

			content, err := os.ReadFile(tc.expectedOutputFile)
			require.NoError(t, err)

			var expected, actual interface{}
			require.NoError(t, yaml.Unmarshal(content, &expected))
			require.NoError(t, yaml.Unmarshal([]byte(output), &actual))
			assert.Equal(t, expected, actual)
		})
	}
}

func Test_FileFormat_InvalidArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		errorString string
	}{
		{
			name:        "invalid type argument",
			args:        []string{"invalid-type", "testdata/file-format/deck-input.yaml"},
			errorString: "invalid value 'invalid-type' found for the 'type' flag",
		},
		{
			name:        "missing file argument",
			args:        []string{"deck"},
			errorString: "accepts 2 arg(s), received 1",
		},
		{
			name:        "no arguments",
			args:        []string{},
			errorString: "accepts 2 arg(s), received 0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fileFormat(tc.args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errorString)
		})
	}
}
