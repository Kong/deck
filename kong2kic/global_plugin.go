package kong2kic

import (
	"encoding/json"
	"log"

	"github.com/kong/go-database-reconciler/pkg/file"
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
		var kongPlugin kicv1.KongClusterPlugin
		kongPlugin.APIVersion = KICAPIVersion
		kongPlugin.Kind = "KongClusterPlugin"
		kongPlugin.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		if plugin.Name != nil {
			kongPlugin.PluginName = *plugin.Name
			kongPlugin.ObjectMeta.Name = calculateSlug(*plugin.Name)
		} else {
			log.Println("Global Plugin name is empty. This is not recommended." +
				"Please, provide a name for the plugin before generating Kong Ingress Controller manifests.")
			continue
		}

		// transform the plugin config from map[string]interface{} to apiextensionsv1.JSON
		var configJSON apiextensionsv1.JSON
		var err error
		configJSON.Raw, err = json.Marshal(plugin.Config)
		if err != nil {
			return err
		}
		kongPlugin.Config = configJSON
		file.KongClusterPlugins = append(file.KongClusterPlugins, kongPlugin)
	}
	return nil
}
