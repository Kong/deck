package kong2kic

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/stretchr/testify/require"
)

var baseLocation = "testdata/"

// Helper function to read and fix JSON stream
func readAndFixJSONStream(filename string) (string, error) {
	content, err := os.ReadFile(filepath.Join(baseLocation, filename))
	if err != nil {
		return "", err
	}
	// Update to an actual JSON array
	fixedContent := "[" + strings.ReplaceAll(string(content), "}{", "},{") + "]"
	return fixedContent, nil
}

// Helper function to compare file content
func compareFileContent(t *testing.T, expectedFilename string, actualContent []byte) {
	expectedPath := filepath.Join(baseLocation, expectedFilename)
	expectedContent, err := os.ReadFile(expectedPath)
	require.NoError(t, err)

	// Write actual content to file for debugging
	actualFilename := strings.Replace(expectedFilename, "-expected.", "-actual.", 1)
	actualPath := filepath.Join(baseLocation, actualFilename)
	err = os.WriteFile(actualPath, actualContent, 0o600)
	require.NoError(t, err)

	if strings.HasSuffix(expectedFilename, ".json") {
		expectedJSON, err := readAndFixJSONStream(expectedFilename)
		require.NoError(t, err)
		actualJSON := "[" + strings.ReplaceAll(string(actualContent), "}{", "},{") + "]"
		require.JSONEq(t, expectedJSON, actualJSON)
	} else {
		// Split the content into individual YAML documents using regex
		re := regexp.MustCompile(`(?m)^---\s*$`)
		expectedYAMLs := re.Split(string(expectedContent), -1)
		actualYAMLs := re.Split(string(actualContent), -1)

		// Ensure both have the same number of YAML documents
		require.Len(t, actualYAMLs, len(expectedYAMLs), "number of YAML documents do not match")

		// Compare each YAML document
		for i := range expectedYAMLs {
			expectedYAML := strings.TrimSpace(expectedYAMLs[i])
			actualYAML := strings.TrimSpace(actualYAMLs[i])
			require.YAMLEq(t, expectedYAML, actualYAML, "YAML document %d does not match", i+1)
		}
	}
}

func Test_convertKongGatewayToKIC(t *testing.T) {
	tests := []struct {
		name           string
		inputFilename  string
		outputFilename string
		builderType    string
		wantErr        bool
	}{
		{
			// Service does not depend on v2 vs v3, or Gateway vs Ingress
			name:           "Kong to KIC: Service",
			inputFilename:  "service-input.yaml",
			outputFilename: "service-output-expected.yaml",
			builderType:    KICV3GATEWAY,
			wantErr:        false,
		},
		{
			// Route to HTTPRoute, Gateway API and KIC v3.
			// In KIC v3 apiVersion: gateway.networking.k8s.io/v1
			name:           "Kong to KIC: Route API GW, KIC v3",
			inputFilename:  "route-input.yaml",
			outputFilename: "route-gw-v3-output-expected.yaml",
			builderType:    KICV3GATEWAY,
			wantErr:        false,
		},
		{
			// Route to HTTPRoute, Gateway API and KIC v2
			// In KIC v2 apiVersion: gateway.networking.k8s.io/v1beta1
			name:           "Kong to KIC: Route API GW, KIC v2",
			inputFilename:  "route-input.yaml",
			outputFilename: "route-gw-v2-output-expected.yaml",
			builderType:    KICV2GATEWAY,
			wantErr:        false,
		},
		{
			// Route to Ingress, Ingress API. Output does not depend on KIC v2 vs v3
			name:           "Kong to KIC: Route Ingress API",
			inputFilename:  "route-input.yaml",
			outputFilename: "route-ingress-output-expected.yaml",
			builderType:    KICV3INGRESS,
			wantErr:        false,
		},
		{
			// Upstream to KongIngress for KIC v2
			name:           "Kong to KIC: Upstream KIC v2",
			inputFilename:  "upstream-input.yaml",
			outputFilename: "upstream-v2-output-expected.yaml",
			builderType:    KICV2GATEWAY,
			wantErr:        false,
		},
		{
			// Upstream to KongUpstreamPolicy for KIC v3
			name:           "Kong to KIC: Upstream KIC v3",
			inputFilename:  "upstream-input.yaml",
			outputFilename: "upstream-v3-output-expected.yaml",
			builderType:    KICV3GATEWAY,
			wantErr:        false,
		},
		{
			// Global Plugin to KongClusterPlugin. Output does not depend on KIC v2 vs v3
			name:           "Kong to KIC: Global Plugin",
			inputFilename:  "global-plugin-input.yaml",
			outputFilename: "global-plugin-output-expected.yaml",
			builderType:    KICV3GATEWAY,
			wantErr:        false,
		},
		{
			// Consumer to KongConsumer. Output depends on KIC v2 vs v3.
			// KIC v2 uses kongCredType for credential type, KIC v3 uses labels
			name:           "Kong to KIC: Consumer KIC v2",
			inputFilename:  "consumer-input.yaml",
			outputFilename: "consumer-v2-output-expected.yaml",
			builderType:    KICV2GATEWAY,
			wantErr:        false,
		},
		{
			// Consumer to KongConsumer. Output depends on KIC v2 vs v3.
			// KIC v2 uses kongCredType for credential type, KIC v3 uses labels
			name:           "Kong to KIC: Consumer KIC v3",
			inputFilename:  "consumer-input.yaml",
			outputFilename: "consumer-v3-output-expected.yaml",
			builderType:    KICV3GATEWAY,
			wantErr:        false,
		},
		{
			// Consumer Group to KongConsumerGroup. Output does not depend on KIC v2 vs v3
			name:           "Kong to KIC: ConsumerGroup",
			inputFilename:  "consumer-group-input.yaml",
			outputFilename: "consumer-group-output-expected.yaml",
			builderType:    KICV3GATEWAY,
			wantErr:        false,
		},
		{
			// Certificate to Secret type: kubernetes.io/tls.
			// Output does not depend on KIC v2 vs v3
			name:           "Kong to KIC: Certificate",
			inputFilename:  "certificate-input.yaml",
			outputFilename: "certificate-output-expected.yaml",
			builderType:    KICV3GATEWAY,
			wantErr:        false,
		},
		{
			// CA Certificate to Secret type: Opaque.
			// Output does not depend on KIC v2 vs v3
			name:           "Kong to KIC: CA Certificate",
			inputFilename:  "ca-certificate-input.yaml",
			outputFilename: "ca-certificate-output-expected.yaml",
			builderType:    KICV3GATEWAY,
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := filepath.Join(baseLocation, tt.inputFilename)
			inputContent, err := file.GetContentFromFiles([]string{inputPath}, false)
			require.NoError(t, err)

			outputFormat := file.YAML
			if strings.HasSuffix(tt.outputFilename, ".json") {
				outputFormat = file.JSON
			}

			output, err := MarshalKongToKIC(inputContent, tt.builderType, outputFormat)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalKongToKIC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				compareFileContent(t, tt.outputFilename, output)
			}
		})
	}
}
