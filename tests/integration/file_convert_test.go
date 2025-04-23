package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				" Allowed values: [kong-gateway kong-gateway-2.x 2.8]",
		},
		{
			name:                        "Invalid destination format",
			convertCmdSourceFormat:      "kong-gateway-2.x",
			convertCmdDestinationFormat: "random-value",
			errorExpected:               true,
			errorString: "invalid value 'random-value' found for the 'to' flag." +
				" Allowed values: [konnect kong-gateway-3.x 3.4]",
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
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorString)

				return
			}

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

func Test_convertKongGateway28xTo34x(t *testing.T) {
	tests := []struct {
		name               string
		inputFile          string
		expectedOutputFile string
		errorExpected      bool
		errorString        string
	}{
		{
			name:               "auto-fixes plugin configuration for 2.8x",
			inputFile:          "testdata/file-convert/002-kong-gateway-28x-to-34x-migration/28x-plugins.yaml",
			errorExpected:      false,
			expectedOutputFile: "testdata/file-convert/002-kong-gateway-28x-to-34x-migration/34x-expected-plugins.yaml",
		},
		{
			name:               "auto-fixes route configuration for 2.8x",
			inputFile:          "testdata/file-convert/002-kong-gateway-28x-to-34x-migration/28x-routes.yaml",
			errorExpected:      false,
			expectedOutputFile: "testdata/file-convert/002-kong-gateway-28x-to-34x-migration/34x-expected-routes.yaml",
		},
	}

	// This is required to create the full testfile names.
	// We are using pre-defined linting rulesets for conversion,
	// and they are relative to the project root.
	// Thus, we are ensuring that the working directory is set to the project root.
	// Otherwise, linting would fail as tests would try to resolve the
	// rulesets relative to the test file location.
	cwd, _ := os.Getwd()
	projectRoot := filepath.Join(cwd, "../../")
	err := os.Chdir(projectRoot)
	require.NoError(t, err)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fileName := filepath.Join(projectRoot, "tests/integration", tc.inputFile)
			output, err := fileConvert(
				"--from", "2.8",
				"--to", "3.4",
				"--input-file", fileName,
			)

			if tc.errorExpected {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorString)

				return
			}

			require.NoError(t, err)
			outputFileName := filepath.Join(projectRoot + "/tests/integration/" + tc.expectedOutputFile)
			expected, err := readFile(outputFileName)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}
