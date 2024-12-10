package kong2kic

import (
	"log"

	"github.com/kong/go-database-reconciler/pkg/file"
	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Function to populate KIC Consumers and their credentials
func populateKICConsumers(content *file.Content, file *KICContent) error {
	for _, consumer := range content.Consumers {
		if consumer.Username == nil {
			log.Println("Consumer username is empty. Please provide a username for the consumer.")
			continue
		}
		username := *consumer.Username
		kongConsumer := configurationv1.KongConsumer{
			TypeMeta: metav1.TypeMeta{
				APIVersion: ConfigurationKongHQv1,
				Kind:       KongConsumerKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        calculateSlug(username),
				Annotations: map[string]string{IngressClass: ClassName},
			},
			Username: username,
		}
		if consumer.CustomID != nil {
			kongConsumer.CustomID = *consumer.CustomID
		}

		// Add tags to annotations
		addTagsToAnnotations(consumer.Tags, kongConsumer.ObjectMeta.Annotations)

		// Populate credentials
		populateKICKeyAuthSecrets(&consumer, &kongConsumer, file)
		populateKICHMACSecrets(&consumer, &kongConsumer, file)
		populateKICJWTAuthSecrets(&consumer, &kongConsumer, file)
		populateKICBasicAuthSecrets(&consumer, &kongConsumer, file)
		populateKICOAuth2CredSecrets(&consumer, &kongConsumer, file)
		populateKICACLGroupSecrets(&consumer, &kongConsumer, file)
		populateKICMTLSAuthSecrets(&consumer, &kongConsumer, file)

		// Handle plugins associated with the consumer
		for _, plugin := range consumer.Plugins {
			kongPlugin, err := createKongPlugin(plugin, username)
			if err != nil {
				return err
			}
			if kongPlugin == nil {
				continue
			}
			file.KongPlugins = append(file.KongPlugins, *kongPlugin)
			addPluginToAnnotations(kongPlugin.ObjectMeta.Name, kongConsumer.ObjectMeta.Annotations)
		}

		file.KongConsumers = append(file.KongConsumers, kongConsumer)
	}

	return nil
}
