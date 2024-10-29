package kong2kic

import (
	"encoding/json"
	"log"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	kicv1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1"
	kicv1beta1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Helper function to populate consumer group plugins
func createConsumerGroupKongPlugin(plugin *kong.ConsumerGroupPlugin, ownerName string) (*kicv1.KongPlugin, error) {
	if plugin.Name == nil {
		log.Println("Plugin name is empty. Please provide a name for the plugin.")
		return nil, nil
	}
	pluginName := *plugin.Name
	kongPlugin := &kicv1.KongPlugin{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "configuration.konghq.com/v1",
			Kind:       "KongPlugin",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        calculateSlug(ownerName + "-" + pluginName),
			Annotations: map[string]string{IngressClass: ClassName},
		},
		PluginName: pluginName,
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

func populateKICConsumerGroups(content *file.Content, kicContent *KICContent) error {
	for _, consumerGroup := range content.ConsumerGroups {
		if consumerGroup.Name == nil {
			log.Println("Consumer group name is empty. Please provide a name for the consumer group.")
			continue
		}
		groupName := *consumerGroup.Name

		kongConsumerGroup := kicv1beta1.KongConsumerGroup{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "configuration.konghq.com/v1beta1",
				Kind:       "KongConsumerGroup",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        calculateSlug(groupName),
				Annotations: map[string]string{IngressClass: ClassName},
			},
		}

		// Add tags to annotations
		addTagsToAnnotations(consumerGroup.Tags, kongConsumerGroup.ObjectMeta.Annotations)

		// Update the ConsumerGroups field of the KongConsumers
		for _, consumer := range consumerGroup.Consumers {
			if consumer.Username == nil {
				log.Println("Consumer username is empty. Please provide a username for the consumer.")
				continue
			}
			username := *consumer.Username
			for idx := range kicContent.KongConsumers {
				if kicContent.KongConsumers[idx].Username == username {
					kicContent.KongConsumers[idx].ConsumerGroups = append(kicContent.KongConsumers[idx].ConsumerGroups, groupName)
				}
			}
		}

		// Handle plugins
		for _, plugin := range consumerGroup.Plugins {
			kongPlugin, err := createConsumerGroupKongPlugin(plugin, groupName)
			if err != nil {
				return err
			}
			if kongPlugin == nil {
				continue
			}

			kicContent.KongPlugins = append(kicContent.KongPlugins, *kongPlugin)

			// Add plugin to kongConsumerGroup annotations
			addPluginToAnnotations(kongPlugin.ObjectMeta.Name, kongConsumerGroup.ObjectMeta.Annotations)
		}

		kicContent.KongConsumerGroups = append(kicContent.KongConsumerGroups, kongConsumerGroup)
	}

	return nil
}
