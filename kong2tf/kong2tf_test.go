package kong2tf

import (
	"os"
	"strings"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var baseLocation = "testdata/"

func fixJSONstream(input string) string {
	// this is a stream of json files, must update to an actual json array
	return "[" + strings.Replace(input, "}{", "},{", -1) + "]"
}

func compareFileContent(t *testing.T, expectedFilename string, actualContent []byte) {
	expected, err := os.ReadFile(baseLocation + expectedFilename)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	actualFilename := baseLocation + strings.Replace(expectedFilename, "-expected.", "-actual.", 1)
	os.WriteFile(actualFilename, actualContent, 0o600)

	// compare the actual content with the expected content
	// both should be the same terraform file
	// ignore differences in whitespace and newlines
	expectedFields := strings.Fields(string(expected))
	actualFields := strings.Fields(string(actualContent))

	require.Equal(t, expectedFields, actualFields)
}

func Test_convertKongGatewayToTerraform(t *testing.T) {
	tests := []struct {
		name           string
		inputFilename  string
		outputFilename string
		wantErr        bool
	}{
		{
			name:           "service",
			inputFilename:  "service-input.yaml",
			outputFilename: "service-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "route",
			inputFilename:  "route-input.yaml",
			outputFilename: "route-output-expected.tf",
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputContent, err := file.GetContentFromFiles([]string{baseLocation + tt.inputFilename}, false)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			output, err := Convert(inputContent)

			if err == nil {
				compareFileContent(t, tt.outputFilename, []byte(output))
			} else if (err != nil) != tt.wantErr {
				t.Errorf("KongToTerraform error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
