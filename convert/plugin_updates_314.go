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

// sslVerifyPlugins is the list of plugins where ssl_verify default
// changed from false to true in Kong Gateway 3.14 (due to the global
// tls_certificate_verify changing from off to on).
var sslVerifyPlugins = map[string]bool{
	acePluginName:                         true,
	acmePluginName:                        true,
	aiAwsGuardrailPluginName:              true,
	aiAzureContentSafetyPluginName:        true,
	aiLlmAsJudgePluginName:                true,
	aiProxyAdvancedPluginName:             true,
	aiRagInjectorPluginName:               true,
	aiRateLimitingAdvancedPluginName:      true,
	aiRequestTransformerPluginName:        true,
	aiResponseTransformerPluginName:       true,
	aiSemanticCachePluginName:             true,
	aiSemanticPromptGuardPluginName:       true,
	aiSemanticResponseGuardPluginName:     true,
	awsLambdaPluginName:                   true,
	azureFunctionsPluginName:              true,
	basicAuthPluginName:                   true,
	confluentPluginName:                   true,
	confluentConsumePluginName:            true,
	datakitPluginName:                     true,
	forwardProxyPluginName:                true,
	graphqlProxyCacheAdvancedPluginName:   true,
	graphqlRateLimitingAdvancedPluginName: true,
	headerCertAuthPluginName:              true,
	httpLogPluginName:                     true,
	jwtSignerPluginName:                   true,
	kafkaConsumePluginName:                true,
	kafkaLogPluginName:                    true,
	kafkaUpstreamPluginName:               true,
	ldapAuthPluginName:                    true,
	ldapAuthAdvancedPluginName:            true,
	mtlsAuthPluginName:                    true,
	opaPluginName:                         true,
	openidConnectPluginName:               true,
	proxyCacheAdvancedPluginName:          true,
	rateLimitingPluginName:                true,
	rateLimitingAdvancedPluginName:        true,
	requestCalloutPluginName:              true,
	responseRateLimitingPluginName:        true,
	samlPluginName:                        true,
	serviceProtectionPluginName:           true,
	tcpLogPluginName:                      true,
	upstreamOauthPluginName:               true,
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

	// Handle ssl_verify default change (false -> true)
	if sslVerifyPlugins[pluginName] {
		if _, exists := plugin.Config["ssl_verify"]; !exists {
			plugin.Config["ssl_verify"] = false
			cprint.UpdatePrintf(
				"Plugin '%s': setting ssl_verify to false "+
					"(old default, 3.14 defaults to true)\n", pluginName)
		}
	}
}
