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
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "convert one service",
			args: args{
				inputFilename:                "custom_resources/yaml/1-service/input.yaml",
				outputFilenameYAMLCRD:        "custom_resources/yaml/1-service/output-expected.yaml",
				outputFilenameJSONCRD:        "custom_resources/json/1-service/output-expected.json",
				outputFilenameYAMLAnnotation: "annotations/yaml/1-service/output-expected.yaml",
				outputFilenameJSONAnnotation: "annotations/json/1-service/output-expected.json",
			},
			wantErr: false,
		},
		{
			name: "convert one service and one route",
			args: args{
				inputFilename:                "custom_resources/yaml/2-service-and-route/input.yaml",
				outputFilenameYAMLCRD:        "custom_resources/yaml/2-service-and-route/output-expected.yaml",
				outputFilenameJSONCRD:        "custom_resources/json/2-service-and-route/output-expected.json",
				outputFilenameYAMLAnnotation: "annotations/yaml/2-service-and-route/output-expected.yaml",
				outputFilenameJSONAnnotation: "annotations/json/2-service-and-route/output-expected.json",
			},
			wantErr: false,
		},
		{
			name: "convert one service with upstream data",
			args: args{
				inputFilename:                "custom_resources/yaml/3-service-and-upstream/input.yaml",
				outputFilenameYAMLCRD:        "custom_resources/yaml/3-service-and-upstream/output-expected.yaml",
				outputFilenameJSONCRD:        "custom_resources/json/3-service-and-upstream/output-expected.json",
				outputFilenameYAMLAnnotation: "annotations/yaml/3-service-and-upstream/output-expected.yaml",
				outputFilenameJSONAnnotation: "annotations/json/3-service-and-upstream/output-expected.json",
			},
			wantErr: false,
		},
		{
			name: "convert one service with upstream and route",
			args: args{
				inputFilename:                "custom_resources/yaml/4-service-route-upstream/input.yaml",
				outputFilenameYAMLCRD:        "custom_resources/yaml/4-service-route-upstream/output-expected.yaml",
				outputFilenameJSONCRD:        "custom_resources/json/4-service-route-upstream/output-expected.json",
				outputFilenameYAMLAnnotation: "annotations/yaml/4-service-route-upstream/output-expected.yaml",
				outputFilenameJSONAnnotation: "annotations/json/4-service-route-upstream/output-expected.json",
			},
			wantErr: false,
		},
		{
			name: "convert one service with upstream, route, acl auth plugin",
			args: args{
				inputFilename:                "custom_resources/yaml/5-service-route-upstream-acl-auth/input.yaml",
				outputFilenameYAMLCRD:        "custom_resources/yaml/5-service-route-upstream-acl-auth/output-expected.yaml",
				outputFilenameJSONCRD:        "custom_resources/json/5-service-route-upstream-acl-auth/output-expected.json",
				outputFilenameYAMLAnnotation: "annotations/yaml/5-service-route-upstream-acl-auth/output-expected.yaml",
				outputFilenameJSONAnnotation: "annotations/json/5-service-route-upstream-acl-auth/output-expected.json",
			},
			wantErr: false,
		},
		{
			name: "convert one service with upstream, route, basic auth plugin",
			args: args{
				inputFilename:                "custom_resources/yaml/6-service-route-upstream-basic-auth/input.yaml",
				outputFilenameYAMLCRD:        "custom_resources/yaml/6-service-route-upstream-basic-auth/output-expected.yaml",
				outputFilenameJSONCRD:        "custom_resources/json/6-service-route-upstream-basic-auth/output-expected.json",
				outputFilenameYAMLAnnotation: "annotations/yaml/6-service-route-upstream-basic-auth/output-expected.yaml",
				outputFilenameJSONAnnotation: "annotations/json/6-service-route-upstream-basic-auth/output-expected.json",
			},
			wantErr: false,
		},
		{
			name: "convert one service with upstream, route, jwt auth plugin",
			args: args{
				inputFilename:                "custom_resources/yaml/7-service-route-upstream-jwt-auth/input.yaml",
				outputFilenameYAMLCRD:        "custom_resources/yaml/7-service-route-upstream-jwt-auth/output-expected.yaml",
				outputFilenameJSONCRD:        "custom_resources/json/7-service-route-upstream-jwt-auth/output-expected.json",
				outputFilenameYAMLAnnotation: "annotations/yaml/7-service-route-upstream-jwt-auth/output-expected.yaml",
				outputFilenameJSONAnnotation: "annotations/json/7-service-route-upstream-jwt-auth/output-expected.json",
			},
			wantErr: false,
		},
		{
			name: "convert one service with upstream, route, key auth plugin",
			args: args{
				inputFilename:                "custom_resources/yaml/8-service-route-upstream-key-auth/input.yaml",
				outputFilenameYAMLCRD:        "custom_resources/yaml/8-service-route-upstream-key-auth/output-expected.yaml",
				outputFilenameJSONCRD:        "custom_resources/json/8-service-route-upstream-key-auth/output-expected.json",
				outputFilenameYAMLAnnotation: "annotations/yaml/8-service-route-upstream-key-auth/output-expected.yaml",
				outputFilenameJSONAnnotation: "annotations/json/8-service-route-upstream-key-auth/output-expected.json",
			},
			wantErr: false,
		},
		{
			name: "convert one service with upstream, route, mtls auth plugin",
			args: args{
				inputFilename:                "custom_resources/yaml/9-service-route-upstream-mtls-auth/input.yaml",
				outputFilenameYAMLCRD:        "custom_resources/yaml/9-service-route-upstream-mtls-auth/output-expected.yaml",
				outputFilenameJSONCRD:        "custom_resources/json/9-service-route-upstream-mtls-auth/output-expected.json",
				outputFilenameYAMLAnnotation: "annotations/yaml/9-service-route-upstream-mtls-auth/output-expected.yaml",
				outputFilenameJSONAnnotation: "annotations/json/9-service-route-upstream-mtls-auth/output-expected.json",
			},
			wantErr: false,
		},
		{
			name: "convert one service with upstream, route, multiple plugin",
			args: args{
				inputFilename:                "custom_resources/yaml/10-multiple-plugins-same-route/input.yaml",
				outputFilenameYAMLCRD:        "custom_resources/yaml/10-multiple-plugins-same-route/output-expected.yaml",
				outputFilenameJSONCRD:        "custom_resources/json/10-multiple-plugins-same-route/output-expected.json",
				outputFilenameYAMLAnnotation: "annotations/yaml/10-multiple-plugins-same-route/output-expected.yaml",
				outputFilenameJSONAnnotation: "annotations/json/10-multiple-plugins-same-route/output-expected.json",
			},
			wantErr: false,
		},
		{
			name: "convert consumer groups",
			args: args{
				inputFilename:                "custom_resources/yaml/11-consumer-group/input.yaml",
				outputFilenameYAMLCRD:        "custom_resources/yaml/11-consumer-group/output-expected.yaml",
				outputFilenameJSONCRD:        "custom_resources/json/11-consumer-group/output-expected.json",
				outputFilenameYAMLAnnotation: "annotations/yaml/11-consumer-group/output-expected.yaml",
				outputFilenameJSONAnnotation: "annotations/json/11-consumer-group/output-expected.json",
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
		})
	}
}
