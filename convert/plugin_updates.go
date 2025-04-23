package convert

import (
	"fmt"

	"github.com/kong/go-database-reconciler/pkg/cprint"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
)

const (
	rateLimitingAdvancedPluginName = "rate-limiting-advanced"
	rlaNamespaceDefaultLength      = 32
	awsLambdaPluginName            = "aws-lambda"
	httpLogPluginName              = "http-log"
	prefunctionPluginName          = "pre-function"
	postfunctionPluginName         = "post-function"
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

func updatePlugins(content *file.Content) {
	for idx := range content.Plugins {
		plugin := &content.Plugins[idx]
		updateLegacyPluginConfig(plugin)
	}

	for _, service := range content.Services {
		for _, plugin := range service.Plugins {
			updateLegacyPluginConfig(plugin)
		}
	}

	for _, route := range content.Routes {
		for _, plugin := range route.Plugins {
			updateLegacyPluginConfig(plugin)
		}
	}

	for _, consumer := range content.Consumers {
		for _, plugin := range consumer.Plugins {
			updateLegacyPluginConfig(plugin)
		}
	}

	for _, consumerGroup := range content.ConsumerGroups {
		for _, plugin := range consumerGroup.Plugins {
			updateLegacyPluginConfig(&file.FPlugin{
				Plugin: kong.Plugin{
					ID:     plugin.ID,
					Name:   plugin.Name,
					Config: plugin.Config,
				},
			})
		}
	}
}

func updateLegacyPluginConfig(plugin *file.FPlugin) {
	if plugin != nil && plugin.Config != nil {
		config := plugin.Config.DeepCopy()

		config = updateLegacyFieldToNewField(config, "blacklist", "deny", "")

		config = updateLegacyFieldToNewField(config, "whitelist", "allow", "")

		if plugin.Name != nil {
			pluginName := *plugin.Name
			if pluginName == awsLambdaPluginName {
				config = removeDeprecatedFields3x(config, "proxy_scheme", pluginName)
			}
			if pluginName == prefunctionPluginName || pluginName == postfunctionPluginName {
				config = updateLegacyFieldToNewField(config, "functions", "access", pluginName)
			}
		}

		plugin.Config = config
	}
}

func updateLegacyFieldToNewField(pluginConfig kong.Configuration,
	oldField, newField, pluginName string,
) kong.Configuration {
	var changes int
	if _, ok := pluginConfig[oldField]; ok {
		pluginConfig[newField] = pluginConfig[oldField]
		delete(pluginConfig, oldField)
		changes++
	}
	if changes > 0 {
		cprint.UpdatePrintf("Automatically converted legacy configuration field \"%s\" to the new field %s in plugin %s\n",
			oldField, newField, pluginName)
	}
	return pluginConfig
}

func removeDeprecatedFields3x(pluginConfig kong.Configuration, fieldName, pluginName string) kong.Configuration {
	var changes int
	if pluginConfig != nil {
		delete(pluginConfig, fieldName)
		changes++
	}
	if changes > 0 {
		cprint.UpdatePrintf("Automatically removed deprecated config field \"%s\" from plugin %s\n", fieldName, pluginName)
	}
	return pluginConfig
}
