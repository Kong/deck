package convert

import (
	"fmt"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
)

const (
	rateLimitingAdvancedPluginName = "rate-limiting-advanced"
	rlaNamespaceDefaultLength      = 32
)

func generateAutoFields(content *file.Content) error {
	for _, plugin := range content.Plugins {
		if *plugin.Name == rateLimitingAdvancedPluginName {
			plugin := plugin
			if err := autoGenerateNamespaceForRLAPlugin(&plugin); err != nil {
				return err
			}
		}
	}

	for _, service := range content.Services {
		for _, plugin := range service.Plugins {
			if *plugin.Name == rateLimitingAdvancedPluginName {
				if err := autoGenerateNamespaceForRLAPlugin(plugin); err != nil {
					return err
				}
			}
		}
	}

	for _, route := range content.Routes {
		for _, plugin := range route.Plugins {
			if *plugin.Name == rateLimitingAdvancedPluginName {
				if err := autoGenerateNamespaceForRLAPlugin(plugin); err != nil {
					return err
				}
			}
		}
	}

	for _, consumer := range content.Consumers {
		for _, plugin := range consumer.Plugins {
			if *plugin.Name == rateLimitingAdvancedPluginName {
				if err := autoGenerateNamespaceForRLAPlugin(plugin); err != nil {
					return err
				}
			}
		}
	}

	for _, consumerGroup := range content.ConsumerGroups {
		for _, plugin := range consumerGroup.Plugins {
			if *plugin.Name == rateLimitingAdvancedPluginName {
				if err := autoGenerateNamespaceForRLAPluginConsumerGroups(plugin); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func autoGenerateNamespaceForRLAPlugin(plugin *file.FPlugin) error {
	if plugin.Config != nil {
		ns, ok := plugin.Config["namespace"]
		if !ok || ns == nil {
			// namespace is not set, generate one.
			randomNamespace, err := randomString(rlaNamespaceDefaultLength)
			if err != nil {
				return fmt.Errorf("error generating random namespace: %w", err)
			}
			plugin.Config["namespace"] = randomNamespace
		}
	}
	return nil
}

func autoGenerateNamespaceForRLAPluginConsumerGroups(plugin *kong.ConsumerGroupPlugin) error {
	if plugin.Config != nil {
		ns, ok := plugin.Config["namespace"]
		if !ok || ns == nil {
			// namespace is not set, generate one.
			randomNamespace, err := randomString(rlaNamespaceDefaultLength)
			if err != nil {
				return fmt.Errorf("error generating random namespace: %w", err)
			}
			plugin.Config["namespace"] = randomNamespace
		}
	}
	return nil
}
