package kong2kic

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters/addons/kong"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters/addons/metallb"
	environment "github.com/kong/kubernetes-testing-framework/pkg/environments"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
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
		require.YAMLEq(t, string(expectedContent), string(actualContent))
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

func Test_deployManifests(t *testing.T) {
	versions := []string{"2.12", "3.0", "3.1", "3.2", "3.3"}
	for _, version := range versions {
		t.Run("KIC Version "+version, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// Configure the testing environment with the specified KIC version
			env, kongAddon, err := setupTestingEnvironmentWithVersion(ctx, version)
			require.NoError(t, err)
			defer teardownEnvironment(ctx, t, env)

			t.Log("waiting for the test environment to be ready for use")
			require.NoError(t, <-env.WaitForReady(ctx))

			t.Log("verifying the test environment becomes ready for use")
			waitForObjects, ready, err := env.Ready(ctx)
			require.NoError(t, err)
			require.Empty(t, waitForObjects)
			require.True(t, ready)

			t.Log("verifying the kong proxy is returning its default 404 response")
			proxyURL, err := getKongProxyURL(ctx, env)
			require.NoError(t, err)
			verifyKongProxyResponse(t, proxyURL)

			t.Log("verifying that the kong addon deployed both proxy and controller")
			verifyKongDeployment(ctx, t, env, kongAddon)

			config := env.Cluster().Config()

			t.Log("deploying the Gateway API CRDs")
			clientset, err := deployGatewayAPICRDs(t, config)
			require.NoError(t, err)

			t.Log("obtaining the ServerPreferredResources from the cluster")
			kindToResource, err := getKindToResourceMap(clientset)
			require.NoError(t, err)

			t.Log("creating a dynamic client for Kubernetes resources")
			dynamicClient, err := dynamic.NewForConfig(config)
			require.NoError(t, err)

			t.Log("deploying manifests to the cluster")
			err = deployManifestsToClusterForVersion(t, dynamicClient, kindToResource, version)
			require.NoError(t, err)
		})
	}
}

// Helper function to set up the testing environment with a specific KIC version
func setupTestingEnvironmentWithVersion(
	ctx context.Context,
	kicVersion string,
) (environment.Environment, *kong.Addon, error) {
	builder := environment.NewBuilder()
	kongAddonBuilder := kong.NewBuilder().
		WithControllerImage("kong/kubernetes-ingress-controller", kicVersion).
		WithProxyImage("kong", "3.4") // Adjust proxy image if needed

	kongAddon := kongAddonBuilder.Build()
	env, err := builder.WithAddons(metallb.New(), kongAddon).Build(ctx)
	if err != nil {
		return nil, nil, err
	}
	return env, kongAddon, nil
}

// Mutex to avoid race condition on ~/.kube/config file
var teardownMutex sync.Mutex

func teardownEnvironment(ctx context.Context, t *testing.T, env environment.Environment) {
	// Lock the mutex to ensure only one teardown process at a time
	teardownMutex.Lock()
	defer teardownMutex.Unlock()

	t.Logf("cleaning up environment %s and cluster %s", env.Name(), env.Cluster().Name())
	require.NoError(t, env.Cleanup(ctx))
}

// Helper function to get Kong proxy URL
func getKongProxyURL(ctx context.Context, env environment.Environment) (string, error) {
	kongAon, err := env.Cluster().GetAddon("kong")
	if err != nil {
		return "", err
	}
	kongAddonRaw, ok := kongAon.(*kong.Addon)
	if !ok {
		return "", errors.New("failed to cast kong addon")
	}
	proxyURL, err := kongAddonRaw.ProxyHTTPURL(ctx, env.Cluster())
	if err != nil {
		return "", err
	}
	return proxyURL.String(), nil
}

// Helper function to verify Kong proxy response
func verifyKongProxyResponse(t *testing.T, proxyURL string) {
	httpc := http.Client{Timeout: time.Second * 10}
	require.Eventually(t, func() bool {
		resp, err := httpc.Get(proxyURL)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusNotFound
	}, time.Minute*3, time.Second)
}

// Helper function to verify Kong deployment
func verifyKongDeployment(ctx context.Context, t *testing.T, env environment.Environment, kongAddon *kong.Addon) {
	client := env.Cluster().Client()
	appsV1 := client.AppsV1()
	deployments := appsV1.Deployments(kongAddon.Namespace())
	kongDeployment, err := deployments.Get(ctx, "ingress-controller-kong", metav1.GetOptions{})
	require.NoError(t, err)
	require.Len(t, kongDeployment.Spec.Template.Spec.Containers, 2)
	require.Equal(t, "ingress-controller", kongDeployment.Spec.Template.Spec.Containers[0].Name)
	require.Equal(t, "proxy", kongDeployment.Spec.Template.Spec.Containers[1].Name)
}

// Helper function to deploy Gateway API CRDs
func deployGatewayAPICRDs(t *testing.T, config *rest.Config) (*clientset.Clientset, error) {
	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	gatewayAPICrdPath := filepath.Join("testdata", "gateway-api-crd.yaml")
	gatewayAPICrdFile, err := os.ReadFile(gatewayAPICrdPath)
	if err != nil {
		return nil, err
	}

	// Split the YAML file into individual documents.
	yamlDocs := regexp.MustCompile(`(?m)^---\s*$`).Split(string(gatewayAPICrdFile), -1)

	for _, doc := range yamlDocs {
		if strings.TrimSpace(doc) == "" {
			continue
		}

		dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(doc)), 4096)
		var crd apiextensionsv1.CustomResourceDefinition
		err := dec.Decode(&crd)
		if err != nil {
			return nil, err
		}

		_, err = clientset.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), &crd, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
		t.Logf("created CRD: %s", crd.Name)
	}

	// Wait for CRDs to be available
	time.Sleep(2 * time.Second)
	return clientset, nil
}

// Helper function to get Kind to Resource mapping
func getKindToResourceMap(clientset *clientset.Clientset) (map[string]string, error) {
	kindToResource := make(map[string]string)
	groups, err := clientset.Discovery().ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		for _, resource := range group.APIResources {
			kindToResource[resource.Kind] = resource.Name
		}
	}
	return kindToResource, nil
}

// Helper function to deploy manifests to the cluster
func deployManifestsToClusterForVersion(
	t *testing.T,
	dynamicClient dynamic.Interface,
	kindToResource map[string]string,
	version string,
) error {
	files, err := os.ReadDir("testdata/")
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := file.Name()
		if !strings.HasSuffix(filename, "output-expected.yaml") {
			continue
		}
		// Skip files based on version
		if version == "2.12" && strings.Contains(filename, "-v3-") {
			continue
		}
		if version != "2.12" && strings.Contains(filename, "-v2-") {
			continue
		}
		content, err := os.ReadFile(filepath.Join("testdata", filename))
		if err != nil {
			return err
		}
		t.Logf("DEPLOYING MANIFEST: %s for KIC version %s", filename, version)
		err = deployManifestToCluster(t, content, kindToResource, dynamicClient)
		if err != nil {
			return err
		}
	}
	return nil
}

// Simplify the deployManifestToCluster function
func deployManifestToCluster(
	t *testing.T,
	manifest []byte,
	kindToResource map[string]string,
	dynamicClient dynamic.Interface,
) error {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 4096)
	var objectsToDelete []ObjectToDelete

	for {
		var rawObj unstructured.Unstructured
		if err := decoder.Decode(&rawObj); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		gvr, err := getGroupVersionResource(&rawObj, kindToResource)
		if err != nil {
			return err
		}

		setNamespaceIfNeeded(&rawObj)

		_, err = dynamicClient.Resource(gvr).
			Namespace(rawObj.GetNamespace()).
			Create(context.TODO(), &rawObj, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		t.Logf("created object: %s of Kind: %s in Namespace: %s", rawObj.GetName(), rawObj.GetKind(), rawObj.GetNamespace())
		objectsToDelete = append(objectsToDelete, ObjectToDelete{object: rawObj, gvr: gvr})
	}

	// Clean up created objects
	for _, obj := range objectsToDelete {
		err := dynamicClient.Resource(obj.gvr).
			Namespace(obj.object.GetNamespace()).
			Delete(context.TODO(), obj.object.GetName(), metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		t.Logf("deleted object: %s of Kind: %s in Namespace: %s",
			obj.object.GetName(),
			obj.object.GetKind(),
			obj.object.GetNamespace())
	}
	return nil
}

// Helper function to get GroupVersionResource from an unstructured object
func getGroupVersionResource(
	obj *unstructured.Unstructured,
	kindToResource map[string]string,
) (schema.GroupVersionResource, error) {
	apiVersion := obj.GetAPIVersion()
	kind := obj.GetKind()
	resource, exists := kindToResource[kind]
	if !exists {
		return schema.GroupVersionResource{}, errors.New("resource not found for kind: " + kind)
	}

	parts := strings.Split(apiVersion, "/")
	if len(parts) == 2 {
		return schema.GroupVersionResource{
			Group:    parts[0],
			Version:  parts[1],
			Resource: resource,
		}, nil
	} else if len(parts) == 1 {
		return schema.GroupVersionResource{
			Group:    "",
			Version:  parts[0],
			Resource: resource,
		}, nil
	}
	return schema.GroupVersionResource{}, errors.New("invalid apiVersion: " + apiVersion)
}

// Helper function to set namespace if needed
func setNamespaceIfNeeded(obj *unstructured.Unstructured) {
	if obj.GetKind() == "KongClusterPlugin" {
		obj.SetNamespace(apiv1.NamespaceAll)
	} else if obj.GetNamespace() == "" {
		obj.SetNamespace(apiv1.NamespaceDefault)
	}
}

// Type definition for objects to delete
type ObjectToDelete struct {
	object unstructured.Unstructured
	gvr    schema.GroupVersionResource
}
