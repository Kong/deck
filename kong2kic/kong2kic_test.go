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
	"testing"
	"time"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters/addons/kong"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters/addons/metallb"
	environment "github.com/kong/kubernetes-testing-framework/pkg/environments"
	"github.com/stretchr/testify/assert"
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

func Test_deployManifests(t *testing.T) {
	var ctx = context.Background()
	t.Parallel()

	t.Log("configuring the testing environment")
	builder := environment.NewBuilder()

	kongAddon := kong.NewBuilder().Build()

	t.Log("building the testing environment and Kubernetes cluster")
	env, err := builder.WithAddons(metallb.New(), kongAddon).Build(ctx)
	require.NoError(t, err)

	t.Logf("setting up the environment cleanup for environment %s and cluster %s", env.Name(), env.Cluster().Name())
	defer func() {
		t.Logf("cleaning up environment %s and cluster %s", env.Name(), env.Cluster().Name())
		require.NoError(t, env.Cleanup(ctx))
	}()

	t.Log("verifying that both addons have been loaded into the environment")
	require.Len(t, env.Cluster().ListAddons(), 2)

	t.Log("waiting for the test environment to be ready for use")
	require.NoError(t, <-env.WaitForReady(ctx))

	t.Log("verifying the test environment becomes ready for use")
	waitForObjects, ready, err := env.Ready(ctx)
	require.NoError(t, err)
	require.Len(t, waitForObjects, 0)
	require.True(t, ready)

	t.Logf("pulling the kong addon from the environment's cluster to verify proxy URL")
	kongAon, err := env.Cluster().GetAddon("kong")
	require.NoError(t, err)
	kongAddonRaw, ok := kongAon.(*kong.Addon)
	require.True(t, ok)
	proxyURL, err := kongAddonRaw.ProxyURL(ctx, env.Cluster())
	require.NoError(t, err)

	t.Log("verifying the kong proxy is returning its default 404 response")
	httpc := http.Client{Timeout: time.Second * 10}
	require.Eventually(t, func() bool {
		resp, err := httpc.Get(proxyURL.String())
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusNotFound
	}, time.Minute*3, time.Second)

	t.Log("verifying that the kong addon deployed both proxy and controller")
	client := env.Cluster().Client()
	appsV1 := client.AppsV1()
	deployments := appsV1.Deployments(kongAddonRaw.Namespace())
	kongDeployment, err := deployments.Get(ctx, "ingress-controller-kong", metav1.GetOptions{})
	require.NoError(t, err)
	require.Len(t, kongDeployment.Spec.Template.Spec.Containers, 2)
	require.Equal(t, kongDeployment.Spec.Template.Spec.Containers[0].Name, "ingress-controller")
	require.Equal(t, kongDeployment.Spec.Template.Spec.Containers[1].Name, "proxy")

	config := env.Cluster().Config()

	// deploy the Gateway API CRDs 
	clientset := deployGatewayAPICRDs(t, config)

	// obtain the ServerPreferredResources from the cluster
	// and create a data structure to retrieve the resource name
	// based on the resource kind
	kindToResource := make(map[string]string)

	t.Log("obtaining the ServerPreferredResources from the cluster")
	groups, err := clientset.Discovery().ServerPreferredResources()
	require.NoError(t, err)
	for _, group := range groups {
		for _, resource := range group.APIResources {
			// Store the resource name in the map, using the Kind as the key
			kindToResource[resource.Kind] = resource.Name
			t.Logf("resource: %s, kind: %s , group: %s, version: %s",
				resource.Name,
				resource.Kind,
				resource.Group,
				resource.Version)
		}
	}

	// Create a dynamic client for kubernetes resources.
	t.Log("creating a dynamic client for kubernetes resources")
	dynamicClient, err := dynamic.NewForConfig(config)
	require.NoError(t, err)

	// iterate over the files in the testdata/ directory
	// ending in output-expected.yaml and deploy them into the cluster
	files, err := os.ReadDir("testdata/")
	require.NoError(t, err)

	for _, file := range files {
		if strings.HasSuffix(file.Name(), "output-expected.yaml") {
			content, err := os.ReadFile(filepath.Join("testdata", file.Name()))
			require.NoError(t, err)
			t.Logf("deploying manifest: %s", file.Name())
			deployManifestToCluster(t, content, kindToResource, dynamicClient)
		}
	}
}

type ObjectToDelete struct {
	object unstructured.Unstructured
	gvr    schema.GroupVersionResource
}

// deploys the resources in the manifest to verify it is valid. Then deletes them.
func deployManifestToCluster(
	t *testing.T,
	manifest []byte,
	kindToResource map[string]string,
	dynamicClient *dynamic.DynamicClient,
) {
	decoder := yaml.NewYAMLOrJSONDecoder(io.NopCloser(bytes.NewReader(manifest)), 4096)
	var rawObj unstructured.Unstructured
	var gvr schema.GroupVersionResource
	var objectsToDelete []ObjectToDelete

	for {

		if err := decoder.Decode(&rawObj); err != nil {
			if errors.Is(err, io.EOF) {
				// End of the manifest
				break
			}
			require.NoError(t, err)
		}

		var group, version, resource string
		if strings.Contains(rawObj.GetAPIVersion(), "/") {
			// if the object APIVersion has "/" then it is a group API, with group and version separated by "/"
			group = strings.Split(rawObj.GetAPIVersion(), "/")[0]
			version = strings.Split(rawObj.GetAPIVersion(), "/")[1]
			resource = kindToResource[rawObj.GetKind()]
		} else {
			// if the object APIVersion does not have "/" then it is a core API with no group, only version
			group = ""
			version = rawObj.GetAPIVersion()
			resource = kindToResource[rawObj.GetKind()]
		}

		gvr = schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: resource,
		}

		if rawObj.GetKind() == "KongClusterPlugin" {
			// KongClusterPlugin is a global resource.
			rawObj.SetNamespace(apiv1.NamespaceAll)
		} else if rawObj.GetNamespace() == apiv1.NamespaceAll {
			// the other resources go to the default namespace
			rawObj.SetNamespace(apiv1.NamespaceDefault)
		}
		_, err := dynamicClient.Resource(gvr).
			Namespace(rawObj.GetNamespace()).
			Create(context.TODO(), &rawObj, metav1.CreateOptions{})
		require.NoError(t, err, "error creating object %s of Kind %s in Namespace %s (APIVersion: %s)",
			rawObj.GetName(),
			rawObj.GetKind(),
			rawObj.GetNamespace(),
			rawObj.GetAPIVersion())
		t.Logf("created object: %s of Kind: %s in Namespace: %s", rawObj.GetName(), rawObj.GetKind(), rawObj.GetNamespace())
		// save the rawObj.GetName() to delete it later in a slice of strings
		objectsToDelete = append(objectsToDelete, ObjectToDelete{object: rawObj, gvr: gvr})
	}
	// delete objects created in the cluster
	for _, objectToDelete := range objectsToDelete {
		err := dynamicClient.Resource(objectToDelete.gvr).
			Namespace(objectToDelete.object.GetNamespace()).
			Delete(context.TODO(), objectToDelete.object.GetName(), metav1.DeleteOptions{})
		require.NoError(t, err, "error deleting object %s of Kind %s in Namespace %s (APIVersion: %s)",
			objectToDelete.object.GetName(),
			objectToDelete.object.GetKind(),
			objectToDelete.object.GetNamespace(),
			objectToDelete.object.GetAPIVersion())
		t.Logf("deleted object: %s of Kind: %s in Namespace: %s",
			objectToDelete.object.GetName(),
			objectToDelete.object.GetKind(),
			objectToDelete.object.GetNamespace())
	}
}

func deployGatewayAPICRDs(t *testing.T, config *rest.Config) *clientset.Clientset {
	t.Log("creating a Kubernetes client to deploy CRDs")
	clientset, err := clientset.NewForConfig(config)
	require.NoError(t, err)

	t.Log("reading Gateway API CRDs file")
	gatewayAPICrdPath := "testdata/gateway-api-crd.yaml"
	gatewayAPICrdFile, err := os.ReadFile(gatewayAPICrdPath)
	require.NoError(t, err)

	// Create a regular expression that matches "---\n" at the beginning of a line
	re := regexp.MustCompile(`(?m)^---\n`)

	// Split the YAML file into individual documents.
	yamlDocs := re.Split(string(gatewayAPICrdFile), -1)

	t.Log("deploying Gateway API CRDs")
	for _, doc := range yamlDocs {

		// Skip empty documents.
		if doc == "" {
			continue
		}

		dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(doc)), 4096)
		var crd apiextensionsv1.CustomResourceDefinition
		err := dec.Decode(&crd)
		require.NoError(t, err)

		_, err = clientset.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), &crd, metav1.CreateOptions{})
		require.NoError(t, err)
		t.Logf("created CRD: %s", crd.Name)
	}

	// Not waiting typically results in (some of) the CRDs not yet being
	// available immediately.
	time.Sleep(1 * time.Second)
	return clientset
}
