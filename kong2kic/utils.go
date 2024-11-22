package kong2kic

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	kcv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func createKongPlugin(plugin *file.FPlugin, ownerName string) (*kcv1.KongPlugin, error) {
	if plugin.Name == nil {
		log.Println("Plugin name is empty. Please provide a name for the plugin.")
		return nil, nil
	}
	pluginName := *plugin.Name
	kongPlugin := &kcv1.KongPlugin{
		TypeMeta: metav1.TypeMeta{
			APIVersion: KICAPIVersion,
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
		kongPlugin.Protocols = kcv1.StringsToKongProtocols(protocols)
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
