package convert

import (
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
)

func updatePluginsFor310(content *file.Content) error {
	for idx := range content.Plugins {
		plugin := &content.Plugins[idx]
		if err := updateLegacyPluginConfigFor310(plugin); err != nil {
			return err
		}
	}

	for _, service := range content.Services {
		for _, plugin := range service.Plugins {
			if err := updateLegacyPluginConfigFor310(plugin); err != nil {
				return err
			}
		}

		for _, route := range service.Routes {
			for _, plugin := range route.Plugins {
				if err := updateLegacyPluginConfigFor310(plugin); err != nil {
					return err
				}
			}
		}
	}

	for _, route := range content.Routes {
		for _, plugin := range route.Plugins {
			if err := updateLegacyPluginConfigFor310(plugin); err != nil {
				return err
			}
		}
	}

	for _, consumer := range content.Consumers {
		for _, plugin := range consumer.Plugins {
			if err := updateLegacyPluginConfigFor310(plugin); err != nil {
				return err
			}
		}
	}

	for _, consumerGroup := range content.ConsumerGroups {
		for _, plugin := range consumerGroup.Plugins {
			if err := updateLegacyPluginConfigFor310(&file.FPlugin{
				Plugin: kong.Plugin{
					ID:     plugin.ID,
					Name:   plugin.Name,
					Config: plugin.Config,
				},
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func updateLegacyPluginConfigFor310(plugin *file.FPlugin) error {
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
			var err error
			config, err = updateLegacyFieldToNewField(config, "redis.cluster_addresses", "redis.cluster_nodes", pluginName)
			if err != nil {
				return err
			}
			config, err = updateLegacyFieldToNewField(config, "redis.sentinel_addresses", "redis.sentinel_nodes", pluginName)
			if err != nil {
				return err
			}
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
			updatedConfig, err := updateLegacyFieldToNewField(
				config, "model.options.upstream_path", "model.options.upstream_url", pluginName)
			if err != nil {
				return err
			}
			config = updatedConfig
		}

		plugin.Config = config
	}
	return nil
}
