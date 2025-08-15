package convert

import (
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
)

func updatePluginsFor310(content *file.Content) {
	for idx := range content.Plugins {
		plugin := &content.Plugins[idx]
		updateLegacyPluginConfigFor310(plugin)
	}

	for _, service := range content.Services {
		for _, plugin := range service.Plugins {
			updateLegacyPluginConfigFor310(plugin)
		}

		for _, route := range service.Routes {
			for _, plugin := range route.Plugins {
				updateLegacyPluginConfigFor310(plugin)
			}
		}
	}

	for _, route := range content.Routes {
		for _, plugin := range route.Plugins {
			updateLegacyPluginConfigFor310(plugin)
		}
	}

	for _, consumer := range content.Consumers {
		for _, plugin := range consumer.Plugins {
			updateLegacyPluginConfigFor310(plugin)
		}
	}

	for _, consumerGroup := range content.ConsumerGroups {
		for _, plugin := range consumerGroup.Plugins {
			updateLegacyPluginConfigFor310(&file.FPlugin{
				Plugin: kong.Plugin{
					ID:     plugin.ID,
					Name:   plugin.Name,
					Config: plugin.Config,
				},
			})
		}
	}
}

func updateLegacyPluginConfigFor310(plugin *file.FPlugin) {
	if plugin != nil && plugin.Config != nil {
		config := plugin.Config.DeepCopy()

		var pluginName string
		if plugin.Name != nil {
			pluginName = *plugin.Name
		}

		_ = pluginName

		// @TODO: Implement based on linting rules
		//config = updateLegacyFieldToNewField(config, "blacklist", "deny", pluginName)

		//config = updateLegacyFieldToNewField(config, "whitelist", "allow", pluginName)

		//if pluginName != "" {
		//	if pluginName == awsLambdaPluginName {
		//		config = removeDeprecatedFields3x(config, "proxy_scheme", pluginName)
		//	}
		//	if pluginName == prefunctionPluginName || pluginName == postfunctionPluginName {
		//		config = updateLegacyFieldToNewField(config, "functions", "access", pluginName)
		//	}
		//}

		plugin.Config = config
	}
}
