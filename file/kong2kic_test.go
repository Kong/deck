package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_convertKongGatewayToIngress(t *testing.T) {
	type args struct {
		inputFilename                string
		outputFilenameYAMLCRD        string
		outputFilenameYAMLAnnotation string
		outputFilenameJSONCRD        string
		outputFilenameJSONAnnotation string
		outputFilenameYAMLGateway    string
		outputFilenameJSONGateway    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Kong to KIC",
			args: args{
				inputFilename:                "input.yaml",
				outputFilenameYAMLCRD:        "custom_resources/yaml/expected-output.yaml",
				outputFilenameJSONCRD:        "custom_resources/json/expected-output.json",
				outputFilenameYAMLAnnotation: "annotations/yaml/expected-output.yaml",
				outputFilenameJSONAnnotation: "annotations/json/expected-output.json",
				outputFilenameYAMLGateway:    "gateway/yaml/expected-output.yaml",
				outputFilenameJSONGateway:    "gateway/json/expected-output.json",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		BaseLocation := "testdata/kong2kic/"
		t.Run(tt.name, func(t *testing.T) {
			inputContent, err := GetContentFromFiles([]string{BaseLocation + tt.args.inputFilename}, false)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			output, err := MarshalKongToKICYaml(inputContent, CUSTOMRESOURCE)
			if (err != nil) != tt.wantErr {
				t.Errorf("KongToKIC() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {

				expected, err := os.ReadFile(BaseLocation + tt.args.outputFilenameYAMLCRD)
				if err != nil {
					assert.Fail(t, err.Error())
				}
				assert.Equal(t, string(expected), string(output))
			}

			output, err = MarshalKongToKICYaml(inputContent, ANNOTATIONS)
			if (err != nil) != tt.wantErr {
				t.Errorf("KongToKIC() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {

				expected, err := os.ReadFile(BaseLocation + tt.args.outputFilenameYAMLAnnotation)
				if err != nil {
					assert.Fail(t, err.Error())
				}
				assert.Equal(t, string(expected), string(output))
			}

			output, err = MarshalKongToKICYaml(inputContent, GATEWAY)
			if (err != nil) != tt.wantErr {
				t.Errorf("KongToKIC() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {

				expected, err := os.ReadFile(BaseLocation + tt.args.outputFilenameYAMLGateway)
				if err != nil {
					assert.Fail(t, err.Error())
				}
				assert.Equal(t, string(expected), string(output))
			}

			output, err = MarshalKongToKICJson(inputContent, CUSTOMRESOURCE)
			if (err != nil) != tt.wantErr {
				t.Errorf("KongToKIC() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {

				expected, err := os.ReadFile(BaseLocation + tt.args.outputFilenameJSONCRD)
				if err != nil {
					assert.Fail(t, err.Error())
				}
				assert.Equal(t, string(expected), string(output))
			}

			output, err = MarshalKongToKICJson(inputContent, ANNOTATIONS)
			if (err != nil) != tt.wantErr {
				t.Errorf("KongToKIC() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {

				expected, err := os.ReadFile(BaseLocation + tt.args.outputFilenameJSONAnnotation)
				if err != nil {
					assert.Fail(t, err.Error())
				}
				assert.Equal(t, string(expected), string(output))
			}

			output, err = MarshalKongToKICJson(inputContent, GATEWAY)
			if (err != nil) != tt.wantErr {
				t.Errorf("KongToKIC() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {

				expected, err := os.ReadFile(BaseLocation + tt.args.outputFilenameJSONGateway)
				if err != nil {
					assert.Fail(t, err.Error())
				}
				assert.Equal(t, string(expected), string(output))
			}
		})
	}
}
