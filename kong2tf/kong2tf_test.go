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

func compareFileContent(t *testing.T, expectedFilename string, actualContent []byte) {
	expected, err := os.ReadFile(baseLocation + expectedFilename)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	actualFilename := baseLocation + strings.Replace(expectedFilename, "-expected.", "-actual.", 1)
	require.NoError(t, os.WriteFile(actualFilename, actualContent, 0o600))

	// compare the actual content with the expected content
	// both should be the same terraform file
	// ignore differences in whitespace and newlines
	expectedFields := strings.Fields(string(expected))
	actualFields := strings.Fields(string(actualContent))

	require.Equal(t, expectedFields, actualFields)
}

func Test_convertKongGatewayToTerraformWithImports(t *testing.T) {
	tests := []struct {
		name           string
		inputFilename  string
		outputFilename string
		wantErr        bool
	}{
		{
			name:           "consumer-jwt",
			inputFilename:  "consumer-jwt-input.yaml",
			outputFilename: "consumer-jwt-output-with-imports-expected.tf",
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputContent, err := file.GetContentFromFiles([]string{baseLocation + tt.inputFilename}, false)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			cpID := new(string)
			*cpID = "abc-123"
			output, err := Convert(inputContent, cpID, true)

			if err == nil {
				compareFileContent(t, tt.outputFilename, []byte(output))
			} else if (err != nil) != tt.wantErr {
				t.Errorf("KongToTerraform error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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
		{
			name:           "ca-certificate",
			inputFilename:  "ca-certificate-input.yaml",
			outputFilename: "ca-certificate-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "certificate",
			inputFilename:  "certificate-sni-input.yaml",
			outputFilename: "certificate-sni-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "consumer-acl",
			inputFilename:  "consumer-acl-input.yaml",
			outputFilename: "consumer-acl-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "consumer-basic-auth",
			inputFilename:  "consumer-basic-auth-input.yaml",
			outputFilename: "consumer-basic-auth-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "consumer-hmac-auth",
			inputFilename:  "consumer-hmac-auth-input.yaml",
			outputFilename: "consumer-hmac-auth-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "consumer-jwt",
			inputFilename:  "consumer-jwt-input.yaml",
			outputFilename: "consumer-jwt-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "consumer-key-auth",
			inputFilename:  "consumer-key-auth-input.yaml",
			outputFilename: "consumer-key-auth-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "consumer-no-auth",
			inputFilename:  "consumer-no-auth-input.yaml",
			outputFilename: "consumer-no-auth-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "consumer-group",
			inputFilename:  "consumer-group-input.yaml",
			outputFilename: "consumer-group-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "consumer-group-plugin",
			inputFilename:  "consumer-group-plugin-input.yaml",
			outputFilename: "consumer-group-plugin-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "consumer-plugin",
			inputFilename:  "consumer-plugin-input.yaml",
			outputFilename: "consumer-plugin-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "global-plugin-oidc",
			inputFilename:  "global-plugin-oidc-input.yaml",
			outputFilename: "global-plugin-oidc-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "global-plugin-rate-limiting",
			inputFilename:  "global-plugin-rate-limiting-input.yaml",
			outputFilename: "global-plugin-rate-limiting-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "route-plugin",
			inputFilename:  "route-plugin-input.yaml",
			outputFilename: "route-plugin-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "service-plugin",
			inputFilename:  "service-plugin-input.yaml",
			outputFilename: "service-plugin-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "upstream",
			inputFilename:  "upstream-target-input.yaml",
			outputFilename: "upstream-target-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "vault",
			inputFilename:  "vault-input.yaml",
			outputFilename: "vault-output-expected.tf",
			wantErr:        false,
		},
		{
			name:           "partial",
			inputFilename:  "partial-input.yaml",
			outputFilename: "partial-output-expected.tf",
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputContent, err := file.GetContentFromFiles([]string{baseLocation + tt.inputFilename}, false)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			output, err := Convert(inputContent, nil, false)

			if err == nil {
				compareFileContent(t, tt.outputFilename, []byte(output))
			} else if (err != nil) != tt.wantErr {
				t.Errorf("KongToTerraform error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
