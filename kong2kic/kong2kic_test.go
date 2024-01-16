package kong2kic

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

	if strings.HasSuffix(expectedFilename, ".json") {
		// this is a stream of json files, must update to an actual json array
		require.JSONEq(t, fixJSONstream(string(expected)), fixJSONstream(string(actualContent)))
	} else {
		require.YAMLEq(t, string(expected), string(actualContent))
	}
}

func Test_convertKongGatewayToIngress(t *testing.T) {
	tests := []struct {
		name           string
		inputFilename  string
		outputFilename string
		builderType    string
		wantErr        bool
	}{
		{
			name:           "Kong to KIC: customresources, yaml",
			inputFilename:  "input.yaml",
			outputFilename: "custom_resources/yaml/output-expected.yaml",
			builderType:    CUSTOMRESOURCE,
			wantErr:        false,
		},
		{
			name:           "Kong to KIC: customresources, json",
			inputFilename:  "input.yaml",
			outputFilename: "custom_resources/json/output-expected.json",
			builderType:    CUSTOMRESOURCE,
			wantErr:        false,
		},
		{
			name:           "Kong to KIC: annotations, yaml",
			inputFilename:  "input.yaml",
			outputFilename: "annotations/yaml/output-expected.yaml",
			builderType:    ANNOTATIONS,
			wantErr:        false,
		},
		{
			name:           "Kong to KIC: annotations, json",
			inputFilename:  "input.yaml",
			outputFilename: "annotations/json/output-expected.json",
			builderType:    ANNOTATIONS,
			wantErr:        false,
		},
		{
			name:           "Kong to KIC: gateway, yaml",
			inputFilename:  "input.yaml",
			outputFilename: "gateway/yaml/output-expected.yaml",
			builderType:    GATEWAY,
			wantErr:        false,
		},
		{
			name:           "Kong to KIC: gateway, json",
			inputFilename:  "input.yaml",
			outputFilename: "gateway/json/output-expected.json",
			builderType:    GATEWAY,
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputContent, err := file.GetContentFromFiles([]string{baseLocation + tt.inputFilename}, false)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			var output []byte
			if strings.HasSuffix(tt.outputFilename, ".json") {
				output, err = MarshalKongToKICJson(inputContent, tt.builderType)
			} else {
				output, err = MarshalKongToKICYaml(inputContent, tt.builderType)
			}

			if err == nil {
				compareFileContent(t, tt.outputFilename, output)
			} else if (err != nil) != tt.wantErr {
				t.Errorf("KongToKIC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
