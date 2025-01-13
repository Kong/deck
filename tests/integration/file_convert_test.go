package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FileConvert(t *testing.T) {
	tests := []struct {
		name                        string
		convertCmdSourceFormat      string
		convertCmdDestinationFormat string
		errorExpected               bool
		errorString                 string
	}{
		{
			name:                        "Valid source and destination format",
			convertCmdSourceFormat:      "kong-gateway-2.x",
			convertCmdDestinationFormat: "kong-gateway-3.x",
			errorExpected:               false,
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
			_, err := fileConvert(
				"--from", tc.convertCmdSourceFormat,
				"--to", tc.convertCmdDestinationFormat,
				"--input-file", "testdata/file-convert/001-kong-gateway-config/kong-gateway-2-x.yaml",
			)

			if tc.errorExpected {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
