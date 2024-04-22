package kong2kic

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/file"
	kicv1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1"
	kicv1beta1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func populateKICConsumerGroups(content *file.Content, kicContent *KICContent) error {
	// iterate over the consumer groups and create a KongConsumerGroup for each one
	for _, consumerGroup := range content.ConsumerGroups {
		var kongConsumerGroup kicv1beta1.KongConsumerGroup
		kongConsumerGroup.APIVersion = "configuration.konghq.com/v1beta1"
		kongConsumerGroup.Kind = "KongConsumerGroup"
		if consumerGroup.Name != nil {
			kongConsumerGroup.Name = *consumerGroup.Name
			kongConsumerGroup.ObjectMeta.Name = calculateSlug(*consumerGroup.Name)
		} else {
			log.Println("Consumer group name is empty. This is not recommended." +
				"Please, provide a name for the consumer group before generating Kong Ingress Controller manifests.")
			continue
		}
		kongConsumerGroup.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}

		// add konghq.com/tags annotation if consumerGroup.Tags is not nil
		if consumerGroup.Tags != nil {
			var tags []string
			for _, tag := range consumerGroup.Tags {
				if tag != nil {
					tags = append(tags, *tag)
				}
			}
			kongConsumerGroup.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
		}

		// Iterate over the consumers in consumerGroup and
		// find the KongConsumer with the same username in kicContent.KongConsumers
		// and add it to the KongConsumerGroup
		for _, consumer := range consumerGroup.Consumers {
			for idx := range kicContent.KongConsumers {
				if kicContent.KongConsumers[idx].Username == *consumer.Username {
					if kicContent.KongConsumers[idx].ConsumerGroups == nil {
						kicContent.KongConsumers[idx].ConsumerGroups = make([]string, 0)
					}
					consumerGroups := append(kicContent.KongConsumers[idx].ConsumerGroups, *consumerGroup.Name)
					kicContent.KongConsumers[idx].ConsumerGroups = consumerGroups
				}
			}
		}

		// for each consumerGroup plugin, create a KongPlugin and a plugin annotation in the kongConsumerGroup
		// to link the plugin. Consumer group plugins are "global plugins" with a consumer_group property
		for _, plugin := range consumerGroup.Plugins {
			var kongPlugin kicv1.KongPlugin
			kongPlugin.APIVersion = KICAPIVersion
			kongPlugin.Kind = KongPluginKind
			kongPlugin.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
			if plugin.Name != nil {
				kongPlugin.PluginName = *plugin.Name
				kongPlugin.ObjectMeta.Name = calculateSlug(*consumerGroup.Name + "-" + *plugin.Name)
			} else {
				log.Println("Plugin name is empty. This is not recommended." +
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
			kicContent.KongPlugins = append(kicContent.KongPlugins, kongPlugin)

			if kongConsumerGroup.ObjectMeta.Annotations["konghq.com/plugins"] == "" {
				kongConsumerGroup.ObjectMeta.Annotations["konghq.com/plugins"] = kongPlugin.ObjectMeta.Name
			} else {
				annotations := kongConsumerGroup.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.ObjectMeta.Name
				kongConsumerGroup.ObjectMeta.Annotations["konghq.com/plugins"] = annotations
			}
		}

		kicContent.KongConsumerGroups = append(kicContent.KongConsumerGroups, kongConsumerGroup)
	}

	return nil
}
