package kong2kic

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	kicv1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func populateKICKongClusterPlugins(content *file.Content, file *KICContent) error {
	// Global Plugins map to KongClusterPlugins
	// iterate content.Plugins and copy them into kicv1.KongPlugin manifests
	// add the kicv1.KongPlugin to the KICContent.KongClusterPlugins slice
	for _, plugin := range content.Plugins {
		// skip this plugin instance if it is a kongconsumergroup plugin.
		// It is a kongconsumergroup plugin if it has a consumer_group property
		if plugin.ConsumerGroup != nil {
			continue
		}
		var kongClusterPlugin kicv1.KongClusterPlugin
		kongClusterPlugin.APIVersion = KICAPIVersion
		kongClusterPlugin.Kind = "KongClusterPlugin"
		kongClusterPlugin.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		if plugin.Name != nil {
			kongClusterPlugin.PluginName = *plugin.Name
			kongClusterPlugin.ObjectMeta.Name = calculateSlug(*plugin.Name)
		} else {
			log.Println("Global Plugin name is empty. This is not recommended." +
				"Please, provide a name for the plugin before generating Kong Ingress Controller manifests.")
			continue
		}

		// populate enabled, runon, ordering and protocols
		if plugin.Enabled != nil {
			kongClusterPlugin.Disabled = !*plugin.Enabled
		}
		if plugin.RunOn != nil {
			kongClusterPlugin.RunOn = *plugin.RunOn
		}
		if plugin.Ordering != nil {
			kongClusterPlugin.Ordering = &kong.PluginOrdering{
				Before: plugin.Ordering.Before,
				After:  plugin.Ordering.After,
			}
		}
		if plugin.Protocols != nil {
			protocols := make([]string, len(plugin.Protocols))
			for i, protocol := range plugin.Protocols {
				if protocol != nil {
					protocols[i] = *protocol
				}
			}
			kongClusterPlugin.Protocols = kicv1.StringsToKongProtocols(protocols)
		}

		// add konghq.com/tags annotation if plugin.Tags is not nil
		if plugin.Tags != nil {
			var tags []string
			for _, tag := range plugin.Tags {
				if tag != nil {
					tags = append(tags, *tag)
				}
			}
			kongClusterPlugin.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
		}

		// transform the plugin config from map[string]interface{} to apiextensionsv1.JSON
		var configJSON apiextensionsv1.JSON
		var err error
		configJSON.Raw, err = json.Marshal(plugin.Config)
		if err != nil {
			return err
		}
		kongClusterPlugin.Config = configJSON
		file.KongClusterPlugins = append(file.KongClusterPlugins, kongClusterPlugin)
	}
	return nil
}
