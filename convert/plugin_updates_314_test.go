package convert

import (
	"testing"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func TestUpdateLegacyPluginConfigFor314_SSLVerifyFields(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *file.FPlugin
		expected kong.Configuration
	}{
		{
			name: "sets nested basic-auth redis ssl_verify",
			plugin: &file.FPlugin{Plugin: kong.Plugin{
				Name: kong.String(basicAuthPluginName),
				Config: kong.Configuration{
					"brute_force_protection": map[string]interface{}{
						"redis": map[string]interface{}{},
					},
				},
			}},
			expected: kong.Configuration{
				"hide_credentials": false,
				"brute_force_protection": map[string]interface{}{
					"redis": map[string]interface{}{
						"ssl_verify": false,
					},
				},
			},
		},
		{
			name: "sets acme redis ssl_verify only when redis config exists",
			plugin: &file.FPlugin{Plugin: kong.Plugin{
				Name: kong.String(acmePluginName),
				Config: kong.Configuration{
					"storage_config": map[string]interface{}{
						"redis": map[string]interface{}{},
					},
				},
			}},
			expected: kong.Configuration{
				"storage_config": map[string]interface{}{
					"redis": map[string]interface{}{
						"ssl_verify": false,
					},
				},
			},
		},
		{
			name: "does not invent missing nested acme config",
			plugin: &file.FPlugin{Plugin: kong.Plugin{
				Name:   kong.String(acmePluginName),
				Config: kong.Configuration{},
			}},
			expected: kong.Configuration{},
		},
		{
			name: "sets openid-connect nested and top-level fields",
			plugin: &file.FPlugin{Plugin: kong.Plugin{
				Name: kong.String(openidConnectPluginName),
				Config: kong.Configuration{
					"cluster_cache_redis": map[string]interface{}{},
					"redis":               map[string]interface{}{},
				},
			}},
			expected: kong.Configuration{
				"ssl_verify":                   false,
				"session_memcached_ssl_verify": false,
				"cluster_cache_redis": map[string]interface{}{
					"ssl_verify": false,
				},
				"redis": map[string]interface{}{
					"ssl_verify": false,
				},
			},
		},
		{
			name: "sets datakit call nodes and redis cache ssl_verify",
			plugin: &file.FPlugin{Plugin: kong.Plugin{
				Name: kong.String(datakitPluginName),
				Config: kong.Configuration{
					"nodes": []interface{}{
						map[string]interface{}{"type": "call", "name": "a"},
						map[string]interface{}{"type": "branch", "name": "b"},
					},
					"resources": map[string]interface{}{
						"cache": map[string]interface{}{
							"redis": map[string]interface{}{},
						},
					},
				},
			}},
			expected: kong.Configuration{
				"nodes": []interface{}{
					map[string]interface{}{"type": "call", "name": "a", "ssl_verify": false},
					map[string]interface{}{"type": "branch", "name": "b"},
				},
				"resources": map[string]interface{}{
					"cache": map[string]interface{}{
						"redis": map[string]interface{}{"ssl_verify": false},
					},
				},
			},
		},
		{
			name: "sets request-callout nested fields",
			plugin: &file.FPlugin{Plugin: kong.Plugin{
				Name: kong.String(requestCalloutPluginName),
				Config: kong.Configuration{
					"cache": map[string]interface{}{
						"redis": map[string]interface{}{},
					},
					"callouts": map[string]interface{}{
						"request": map[string]interface{}{
							"http_opts": map[string]interface{}{},
						},
					},
				},
			}},
			expected: kong.Configuration{
				"cache": map[string]interface{}{
					"redis": map[string]interface{}{"ssl_verify": false},
				},
				"callouts": map[string]interface{}{
					"request": map[string]interface{}{
						"http_opts": map[string]interface{}{"ssl_verify": false},
					},
				},
			},
		},
		{
			name: "uses plugin-specific field names for azure and ldap",
			plugin: &file.FPlugin{Plugin: kong.Plugin{
				Name:   kong.String(azureFunctionsPluginName),
				Config: kong.Configuration{},
			}},
			expected: kong.Configuration{
				"https_verify": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateLegacyPluginConfigFor314(tt.plugin)
			assert.Equal(t, tt.expected, tt.plugin.Config)
		})
	}
}

func TestUpdateLegacyPluginConfigFor314_LDAPVerifyHost(t *testing.T) {
	plugin := &file.FPlugin{Plugin: kong.Plugin{
		Name:   kong.String(ldapAuthPluginName),
		Config: kong.Configuration{},
	}}

	updateLegacyPluginConfigFor314(plugin)

	assert.Equal(t, kong.Configuration{
		"hide_credentials": false,
		"verify_ldap_host": false,
	}, plugin.Config)
}

func TestUpdateLegacyPluginConfigFor314_LeavesUnsupportedPluginUnchanged(t *testing.T) {
	plugin := &file.FPlugin{Plugin: kong.Plugin{
		Name:   kong.String(aiLlmAsJudgePluginName),
		Config: kong.Configuration{"foo": "bar"},
	}}

	updateLegacyPluginConfigFor314(plugin)

	assert.Equal(t, kong.Configuration{"foo": "bar"}, plugin.Config)
}
