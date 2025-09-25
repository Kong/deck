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

		isRedisPlugin := false
		redisPlugins := []string{
			rateLimitingAdvancedPluginName,
			aiRateLimitingAdvancedPluginName,
			graphqlProxyCacheAdvancedPluginName,
			graphqlRateLimitingAdvancedPluginName,
			proxyCacheAdvancedPluginName,
		}

		for _, name := range redisPlugins {
			if pluginName == name {
				isRedisPlugin = true
				break
			}
		}
		if isRedisPlugin {
			config = updateLegacyFieldToNewField(config, "redis.cluster_addresses", "redis.cluster_nodes", pluginName)
			config = updateLegacyFieldToNewField(config, "redis.sentinel_addresses", "redis.sentinel_nodes", pluginName)
		}

		if pluginName == aiRateLimitingAdvancedPluginName {
			config = convertScalarToList(config, "llm_providers.window_size", pluginName)
			config = convertScalarToList(config, "llm_providers.limit", pluginName)
		}

		modelSchemaPlugins := []string{
			aiProxyPluginName,
			aiProxyAdvancedPluginName,
			aiRagInjectorPluginName,
			aiRequestTransformerPluginName,
			aiResponseTransformerPluginName,
			aiSemanticCachePluginName,
			aiSemanticPromptGuardPluginName,
		}

		isModelSchemaPlugin := false
		for _, name := range modelSchemaPlugins {
			if pluginName == name {
				isModelSchemaPlugin = true
				break
			}
		}

		if isModelSchemaPlugin {
			config = updateLegacyFieldToNewField(config, "model.options.upstream_path", "model.options.upstream_url", pluginName)
		}

		plugin.Config = config
	}
}
