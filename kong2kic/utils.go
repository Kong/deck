package kong2kic

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KongHQClientCert              = "konghq.com/client-cert"
	KongHQConnectTimeout          = "konghq.com/connect-timeout"
	KongHQCredential              = "konghq.com/credential" //nolint: gosec
	KongHQHeaders                 = "konghq.com/headers"
	KongHQHTTPSRedirectStatusCode = "konghq.com/https-redirect-status-code"
	KongHQMethods                 = "konghq.com/methods"
	KongHQOverride                = "konghq.com/override"
	KongHQPath                    = "konghq.com/path"
	KongHQPathHandling            = "konghq.com/path-handling"
	KongHQPlugins                 = "konghq.com/plugins"
	KongHQPreserveHost            = "konghq.com/preserve-host"
	KongHQProtocol                = "konghq.com/protocol"
	KongHQProtocols               = "konghq.com/protocols"
	KongHQReadTimeout             = "konghq.com/read-timeout"
	KongHQRegexPriority           = "konghq.com/regex-priority"
	KongHQRequestBuffering        = "konghq.com/request-buffering"
	KongHQResponseBuffering       = "konghq.com/response-buffering"
	KongHQRetries                 = "konghq.com/retries"
	KongHQSNIs                    = "konghq.com/snis"
	KongHQStripPath               = "konghq.com/strip-path"
	KongHQTags                    = "konghq.com/tags"
	KongHQUpstreamPolicy          = "konghq.com/upstream-policy"
	KongHQWriteTimeout            = "konghq.com/write-timeout"
)

const (
	ConfigurationKongHQ        = "configuration.konghq.com"
	ConfigurationKongHQv1      = "configuration.konghq.com/v1"
	ConfigurationKongHQv1beta1 = "configuration.konghq.com/v1beta1"
	GatewayAPIVersionV1        = "gateway.networking.k8s.io/v1"
	GatewayAPIVersionV1Beta1   = "gateway.networking.k8s.io/v1beta1"
	HTTPRouteKind              = "HTTPRoute"
	IngressAPIVersion          = "networking.k8s.io/v1"
	IngressClass               = "kubernetes.io/ingress.class"
	IngressKind                = "Ingress"
	KICV2GATEWAY               = "KICV2_GATEWAY"
	KICV2INGRESS               = "KICV2_INGRESS"
	KICV3GATEWAY               = "KICV3_GATEWAY"
	KICV3INGRESS               = "KICV3_INGRESS"
	KongClusterPluginKind      = "KongClusterPlugin"
	KongConsumerKind           = "KongConsumer"
	KongConsumerGroupKind      = "KongConsumerGroup"
	KongCredType               = "kongCredType"
	KongIngressKind            = "KongIngress"
	KongPluginKind             = "KongPlugin"
	SecretKind                 = "Secret"
	SecretCADigest             = "ca.digest"
	ServiceAPIVersionv1        = "v1"
	ServiceKind                = "Service"
	UpstreamPolicyKind         = "KongUpstreamPolicy"
)

// Helper function to add tags to annotations
func addTagsToAnnotations(tags []*string, annotations map[string]string) {
	if tags != nil {
		var tagList []string
		for _, tag := range tags {
			if tag != nil {
				tagList = append(tagList, *tag)
			}
		}
		if len(tagList) > 0 {
			annotations["konghq.com/tags"] = strings.Join(tagList, ",")
		}
	}
}

// Helper function to update the "konghq.com/plugins" annotation
func addPluginToAnnotations(pluginName string, annotations map[string]string) {
	if existing, ok := annotations["konghq.com/plugins"]; ok && existing != "" {
		annotations["konghq.com/plugins"] = existing + "," + pluginName
	} else {
		annotations["konghq.com/plugins"] = pluginName
	}
}

// Helper function to create a KongPlugin from a plugin
func createKongPlugin(plugin *file.FPlugin, ownerName string) (*configurationv1.KongPlugin, error) {
	if plugin.Name == nil {
		log.Println("Plugin name is empty. Please provide a name for the plugin.")
		return nil, nil
	}
	pluginName := *plugin.Name
	kongPlugin := &configurationv1.KongPlugin{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ConfigurationKongHQv1,
			Kind:       KongPluginKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        calculateSlug(ownerName + "-" + pluginName),
			Annotations: map[string]string{IngressClass: ClassName},
		},
		PluginName: pluginName,
	}

	// Add tags to annotations
	addTagsToAnnotations(plugin.Tags, kongPlugin.ObjectMeta.Annotations)

	// Populate enabled, runon, ordering, and protocols
	if plugin.Enabled != nil {
		kongPlugin.Disabled = !*plugin.Enabled
	}
	if plugin.RunOn != nil {
		kongPlugin.RunOn = *plugin.RunOn
	}
	if plugin.Ordering != nil {
		kongPlugin.Ordering = &kong.PluginOrdering{
			Before: plugin.Ordering.Before,
			After:  plugin.Ordering.After,
		}
	}
	if plugin.Protocols != nil {
		var protocols []string
		for _, protocol := range plugin.Protocols {
			if protocol != nil {
				protocols = append(protocols, *protocol)
			}
		}
		kongPlugin.Protocols = configurationv1.StringsToKongProtocols(protocols)
	}

	// Transform the plugin config
	configJSON, err := json.Marshal(plugin.Config)
	if err != nil {
		return nil, err
	}
	kongPlugin.Config = apiextensionsv1.JSON{
		Raw: configJSON,
	}

	return kongPlugin, nil
}
