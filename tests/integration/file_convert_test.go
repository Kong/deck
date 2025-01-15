package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func Test_FileConvert(t *testing.T) {
	tests := []struct {
		name                        string
		convertCmdSourceFormat      string
		convertCmdDestinationFormat string
		errorExpected               bool
		errorString                 string
		expectedOutputFile          string
	}{
		{
			name:                        "Valid source and destination format",
			convertCmdSourceFormat:      "kong-gateway-2.x",
			convertCmdDestinationFormat: "kong-gateway-3.x",
			errorExpected:               false,
			expectedOutputFile:          "testdata/file-convert/001-kong-gateway-config/kong-gateway-3-x.yaml",
		},
		{
			name:                        "Invalid source format",
			convertCmdSourceFormat:      "random-value",
			convertCmdDestinationFormat: "kong-gateway-3.x",
			errorExpected:               true,
			errorString: "invalid value 'random-value' found for the 'from' flag." +
				" Allowed values: [kong-gateway kong-gateway-2.x]",
		},
		{
			name:                        "Invalid destination format",
			convertCmdSourceFormat:      "kong-gateway-2.x",
			convertCmdDestinationFormat: "random-value",
			errorExpected:               true,
			errorString: "invalid value 'random-value' found for the 'to' flag." +
				" Allowed values: [konnect kong-gateway-3.x]",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := fileConvert(
				"--from", tc.convertCmdSourceFormat,
				"--to", tc.convertCmdDestinationFormat,
				"--input-file", "testdata/file-convert/001-kong-gateway-config/kong-gateway-2-x.yaml",
			)

			if tc.errorExpected {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorString)

				return
			}

			assert.NoError(t, err)

			var expectedOutput interface{}
			var currentOutput interface{}
			content, err := os.ReadFile(tc.expectedOutputFile)
			assert.NoError(t, err)

			err = yaml.Unmarshal(content, &expectedOutput)
			assert.NoError(t, err)
			err = yaml.Unmarshal([]byte(output), &currentOutput)
			assert.NoError(t, err)
			assert.Equal(t, expectedOutput, currentOutput)
		})
	}
}
