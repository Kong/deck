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
			name:           "Kong to KIC: kic v3.x Gateway API, yaml",
			inputFilename:  "input.yaml",
			outputFilename: "kicv3_gateway/yaml/output-expected.yaml",
			builderType:    KICV3GATEWAY,
			wantErr:        false,
		},
		{
			name:           "Kong to KIC: kic v3.x Gateway API, json",
			inputFilename:  "input.yaml",
			outputFilename: "kicv3_gateway/json/output-expected.json",
			builderType:    KICV3GATEWAY,
			wantErr:        false,
		},
		{
			name:           "Kong to KIC: kic v3.x Ingress API, yaml",
			inputFilename:  "input.yaml",
			outputFilename: "kicv3_ingress/yaml/output-expected.yaml",
			builderType:    KICV3INGRESS,
			wantErr:        false,
		},
		{
			name:           "Kong to KIC: kic v3.x Ingress API, json",
			inputFilename:  "input.yaml",
			outputFilename: "kicv3_ingress/json/output-expected.json",
			builderType:    KICV3INGRESS,
			wantErr:        false,
		},
		{
			name:           "Kong to KIC: kic v2.x Gateway API, yaml",
			inputFilename:  "input.yaml",
			outputFilename: "kicv2_gateway/yaml/output-expected.yaml",
			builderType:    KICV2GATEWAY,
			wantErr:        false,
		},
		{
			name:           "Kong to KIC: kic v2.x Gateway API, json",
			inputFilename:  "input.yaml",
			outputFilename: "kicv2_gateway/json/output-expected.json",
			builderType:    KICV2GATEWAY,
			wantErr:        false,
		},
		{
			name:           "Kong to KIC: kic v2.x Ingress API, yaml",
			inputFilename:  "input.yaml",
			outputFilename: "kicv2_ingress/yaml/output-expected.yaml",
			builderType:    KICV2INGRESS,
			wantErr:        false,
		},
		{
			name:           "Kong to KIC: kic v2.x Ingress API, json",
			inputFilename:  "input.yaml",
			outputFilename: "kicv2_ingress/json/output-expected.json",
			builderType:    KICV2INGRESS,
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
				output, err = MarshalKongToKIC(inputContent, tt.builderType, file.JSON)
			} else {
				output, err = MarshalKongToKIC(inputContent, tt.builderType, file.YAML)
			}

			if err == nil {
				compareFileContent(t, tt.outputFilename, output)
			} else if (err != nil) != tt.wantErr {
				t.Errorf("KongToKIC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
