package convert

import (
	"github.com/kong/go-database-reconciler/pkg/cprint"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
)

// hideCredentialsPlugins is the list of auth plugins where hide_credentials
// default changed from false to true in Kong Gateway 3.14.
var hideCredentialsPlugins = map[string]bool{
	keyAuthPluginName:             true,
	keyAuthEncPluginName:          true,
	basicAuthPluginName:           true,
	hmacAuthPluginName:            true,
	ldapAuthPluginName:            true,
	oauth2PluginName:              true,
	oauth2IntrospectionPluginName: true,
	vaultAuthPluginName:           true,
	ldapAuthAdvancedPluginName:    true,
}

type pluginConfigDefaultSetter struct {
	fieldName string
	set       func(kong.Configuration) bool
}

// sslVerifyPluginConfigSetters maps plugins to the specific config fields whose
// TLS verification defaults changed in Kong Gateway 3.14.
var sslVerifyPluginConfigSetters = map[string][]pluginConfigDefaultSetter{
	acePluginName: {
		newNestedBoolDefaultSetter("rate_limiting.redis.ssl_verify"),
	},
	acmePluginName: {
		newNestedBoolDefaultSetter("storage_config.redis.ssl_verify"),
	},
	aiAwsGuardrailPluginName: {
		newNestedBoolDefaultSetter("ssl_verify"),
	},
	aiAzureContentSafetyPluginName: {
		newNestedBoolDefaultSetter("ssl_verify"),
	},
	aiProxyAdvancedPluginName: aiVectorDBSSLVerifySetters(),
	aiRagInjectorPluginName:   aiVectorDBSSLVerifySetters(),
	aiRateLimitingAdvancedPluginName: {
		newNestedBoolDefaultSetter("redis.ssl_verify"),
	},
	aiSemanticCachePluginName:         aiVectorDBSSLVerifySetters(),
	aiSemanticPromptGuardPluginName:   aiVectorDBSSLVerifySetters(),
	aiSemanticResponseGuardPluginName: aiVectorDBSSLVerifySetters(),
	awsLambdaPluginName: {
		newNestedBoolDefaultSetter("ssl_verify"),
	},
	azureFunctionsPluginName: {
		newNestedBoolDefaultSetter("https_verify"),
	},
	basicAuthPluginName: {
		newNestedBoolDefaultSetter("brute_force_protection.redis.ssl_verify"),
	},
	confluentPluginName: {
		newNestedBoolDefaultSetter("security.ssl_verify"),
		newNestedBoolDefaultSetter("schema_registry.confluent.authentication.oauth2_client.ssl_verify"),
	},
	confluentConsumePluginName: {
		newNestedBoolDefaultSetter("security.ssl_verify"),
		newNestedBoolDefaultSetter("schema_registry.confluent.authentication.oauth2_client.ssl_verify"),
	},
	datakitPluginName: {
		newDatakitNodeSSLVerifySetter(),
		newNestedBoolDefaultSetter("resources.cache.redis.ssl_verify"),
	},
	forwardProxyPluginName: {
		newNestedBoolDefaultSetter("https_verify"),
	},
	graphqlProxyCacheAdvancedPluginName: {
		newNestedBoolDefaultSetter("redis.ssl_verify"),
	},
	graphqlRateLimitingAdvancedPluginName: {
		newNestedBoolDefaultSetter("redis.ssl_verify"),
	},
	headerCertAuthPluginName: {
		newNestedBoolDefaultSetter("ssl_verify"),
	},
	httpLogPluginName: {
		newNestedBoolDefaultSetter("ssl_verify"),
	},
	jwtSignerPluginName: {
		newNestedBoolDefaultSetter("access_token_endpoints_ssl_verify"),
		newNestedBoolDefaultSetter("channel_token_endpoints_ssl_verify"),
	},
	kafkaConsumePluginName: {
		newNestedBoolDefaultSetter("security.ssl_verify"),
		newNestedBoolDefaultSetter("schema_registry.confluent.authentication.oauth2_client.ssl_verify"),
		newNestedBoolDefaultSetter("topics.schema_registry.confluent.authentication.oauth2_client.ssl_verify"),
	},
	kafkaLogPluginName: {
		newNestedBoolDefaultSetter("security.ssl_verify"),
		newNestedBoolDefaultSetter("schema_registry.confluent.authentication.oauth2_client.ssl_verify"),
	},
	kafkaUpstreamPluginName: {
		newNestedBoolDefaultSetter("security.ssl_verify"),
		newNestedBoolDefaultSetter("schema_registry.confluent.authentication.oauth2_client.ssl_verify"),
	},
	ldapAuthPluginName: {
		newNestedBoolDefaultSetter("verify_ldap_host"),
	},
	ldapAuthAdvancedPluginName: {
		newNestedBoolDefaultSetter("verify_ldap_host"),
	},
	mtlsAuthPluginName: {
		newNestedBoolDefaultSetter("ssl_verify"),
	},
	openidConnectPluginName: {
		newNestedBoolDefaultSetter("ssl_verify"),
		newNestedBoolDefaultSetter("cluster_cache_redis.ssl_verify"),
		newNestedBoolDefaultSetter("redis.ssl_verify"),
		newNestedBoolDefaultSetter("session_memcached_ssl_verify"),
	},
	proxyCacheAdvancedPluginName: {
		newNestedBoolDefaultSetter("redis.ssl_verify"),
	},
	rateLimitingPluginName: {
		newNestedBoolDefaultSetter("redis.ssl_verify"),
	},
	rateLimitingAdvancedPluginName: {
		newNestedBoolDefaultSetter("redis.ssl_verify"),
	},
	redisPartialsPluginName: {
		newNestedBoolDefaultSetter("ssl_verify"),
	},
	requestCalloutPluginName: {
		newNestedBoolDefaultSetter("cache.redis.ssl_verify"),
		newNestedBoolDefaultSetter("callouts.request.http_opts.ssl_verify"),
	},
	responseRateLimitingPluginName: {
		newNestedBoolDefaultSetter("redis.ssl_verify"),
	},
	samlPluginName: {
		newNestedBoolDefaultSetter("redis.ssl_verify"),
	},
	serviceProtectionPluginName: {
		newNestedBoolDefaultSetter("redis.ssl_verify"),
	},
	tcpLogPluginName: {
		newNestedBoolDefaultSetter("ssl_verify"),
	},
	upstreamOauthPluginName: {
		newNestedBoolDefaultSetter("client.ssl_verify"),
		newNestedBoolDefaultSetter("cache.redis.ssl_verify"),
	},
}

func updatePluginsFor314(content *file.Content) {
	for idx := range content.Plugins {
		plugin := &content.Plugins[idx]
		updateLegacyPluginConfigFor314(plugin)
	}

	for _, service := range content.Services {
		for _, plugin := range service.Plugins {
			updateLegacyPluginConfigFor314(plugin)
		}

		for _, route := range service.Routes {
			for _, plugin := range route.Plugins {
				updateLegacyPluginConfigFor314(plugin)
			}
		}
	}

	for _, route := range content.Routes {
		for _, plugin := range route.Plugins {
			updateLegacyPluginConfigFor314(plugin)
		}
	}

	for _, consumer := range content.Consumers {
		for _, plugin := range consumer.Plugins {
			updateLegacyPluginConfigFor314(plugin)
		}
	}

	for _, consumerGroup := range content.ConsumerGroups {
		for _, plugin := range consumerGroup.Plugins {
			updateLegacyPluginConfigFor314(&file.FPlugin{
				Plugin: kong.Plugin{
					ID:     plugin.ID,
					Name:   plugin.Name,
					Config: plugin.Config,
				},
			})
		}
	}
}

func updateLegacyPluginConfigFor314(plugin *file.FPlugin) {
	if plugin == nil || plugin.Name == nil {
		return
	}

	pluginName := *plugin.Name

	// Initialize config map if nil
	if plugin.Config == nil {
		plugin.Config = kong.Configuration{}
	}

	// Handle hide_credentials default change (false -> true)
	if hideCredentialsPlugins[pluginName] {
		if _, exists := plugin.Config["hide_credentials"]; !exists {
			plugin.Config["hide_credentials"] = false
			cprint.UpdatePrintf(
				"Plugin '%s': setting hide_credentials to false "+
					"(old default, 3.14 defaults to true)\n", pluginName)
		}
	}

	// Handle plugin-specific TLS verification default changes (false -> true)
	if setters, ok := sslVerifyPluginConfigSetters[pluginName]; ok {
		for _, setter := range setters {
			if setter.set(plugin.Config) {
				cprint.UpdatePrintf(
					"Plugin '%s': setting %s to false "+
						"(old default, 3.14 defaults to true)\n", pluginName, setter.fieldName)
			}
		}
	}
}

func aiVectorDBSSLVerifySetters() []pluginConfigDefaultSetter {
	return []pluginConfigDefaultSetter{
		newNestedBoolDefaultSetter("vectordb.pgvector.ssl_verify"),
		newNestedBoolDefaultSetter("vectordb.redis.ssl_verify"),
	}
}

func newNestedBoolDefaultSetter(path string) pluginConfigDefaultSetter {
	return pluginConfigDefaultSetter{
		fieldName: path,
		set: func(config kong.Configuration) bool {
			return setNestedBoolDefault(config, splitPath(path))
		},
	}
}

func newDatakitNodeSSLVerifySetter() pluginConfigDefaultSetter {
	return pluginConfigDefaultSetter{
		fieldName: "nodes[].ssl_verify",
		set: func(config kong.Configuration) bool {
			nodes, ok := config["nodes"].([]interface{})
			if !ok {
				return false
			}

			updated := false
			for _, rawNode := range nodes {
				node, ok := asConfigMap(rawNode)
				if !ok || node["type"] != "call" {
					continue
				}

				if _, exists := node["ssl_verify"]; exists {
					continue
				}

				node["ssl_verify"] = false
				updated = true
			}

			return updated
		},
	}
}

func setNestedBoolDefault(config kong.Configuration, path []string) bool {
	if len(path) == 0 {
		return false
	}

	current := map[string]interface{}(config)
	for _, segment := range path[:len(path)-1] {
		next, exists := current[segment]
		if !exists || next == nil {
			return false
		}

		child, ok := asConfigMap(next)
		if !ok {
			return false
		}

		current = child
	}

	leaf := path[len(path)-1]
	if _, exists := current[leaf]; exists {
		return false
	}

	current[leaf] = false
	return true
}

func asConfigMap(value interface{}) (map[string]interface{}, bool) {
	switch v := value.(type) {
	case kong.Configuration:
		return map[string]interface{}(v), true
	case map[string]interface{}:
		return v, true
	default:
		return nil, false
	}
}

func splitPath(path string) []string {
	segments := make([]string, 0, len(path))
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] != '.' {
			continue
		}

		segments = append(segments, path[start:i])
		start = i + 1
	}

	return append(segments, path[start:])
}
