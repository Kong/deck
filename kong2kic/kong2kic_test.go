package kong2kic

import (
	"os"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/stretchr/testify/assert"
)

var baseLocation = "testdata/"

func compareFileContent(t *testing.T, expectedFilename string, actual []byte) {
	expected, err := os.ReadFile(baseLocation + expectedFilename)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, string(expected), string(actual))
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

			output, err := MarshalKongToKICYaml(inputContent, tt.builderType)
			if err == nil {
				compareFileContent(t, tt.outputFilename, output)
			} else if (err != nil) != tt.wantErr {
				t.Errorf("KongToKIC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
