package sanitize

import (
	"context"
	"reflect"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Sanitize(t *testing.T) {
	fileContent := &file.Content{
		FormatVersion: "3.0",
		Workspace:     "test-workspace",
		Services: []file.FService{
			{
				Service: kong.Service{
					Name:    kong.String("test-service"),
					Host:    kong.String("localhost"),
					Port:    kong.Int(8000),
					Path:    kong.String("/api"),
					Enabled: kong.Bool(true),
					Tags: []*string{
						kong.String("tag1"),
						kong.String("tag2"),
					},
				},
			},
		},
		Routes: []file.FRoute{
			{
				Route: kong.Route{
					Name: kong.String("test-route"),
					Paths: []*string{
						kong.String("/testOne"),
						kong.String("/testTwo"),
					},
					Service: &kong.Service{
						Name: kong.String("test-service"),
					},
				},
			},
		},
	}

	sanitizer := NewSanitizer(&SanitizerOptions{
		Ctx: context.Background(),
		// Using DeepCopy to avoid modifying the original content for tests
		Content: fileContent.DeepCopy(),
	})

	result, err := sanitizer.Sanitize()
	require.NoError(t, err)
	require.NotNil(t, result)
	t.Run("doesn't sanitize top-level exempted fields", func(t *testing.T) {
		assert.Equal(t, fileContent.FormatVersion, result.FormatVersion, "FormatVersion field should remain unchanged")
		assert.Equal(t, fileContent.Workspace, result.Workspace, "Workspace field should remain unchanged")
	})

	t.Run("sanitizes only string fields", func(t *testing.T) {
		// ensuring that result has all expected fields
		require.NotNil(t, result.Services)
		require.NotNil(t, result.Routes)

		require.Len(t, result.Services, 1)
		require.Len(t, result.Routes, 1)

		require.NotNil(t, result.Services[0].Name)
		require.NotNil(t, result.Services[0].Host)
		require.NotNil(t, result.Services[0].Path)
		require.NotNil(t, result.Routes[0].Name)
		require.Len(t, result.Routes[0].Paths, 2)

		// Verifying that sanitized values are different from original
		// Service is sanitized
		assert.NotEqual(t, *fileContent.Services[0].Name, *result.Services[0].Name)
		assert.NotEqual(t, *fileContent.Services[0].Host, *result.Services[0].Host)
		assert.NotEqual(t, *fileContent.Services[0].Path, *result.Services[0].Path)
		for i, tag := range result.Services[0].Tags {
			assert.NotEqual(t, *fileContent.Services[0].Tags[i], *tag)
		}

		// Route is sanitized
		assert.NotEqual(t, *fileContent.Routes[0].Name, *result.Routes[0].Name)
		for i, path := range result.Routes[0].Paths {
			assert.NotEqual(t, *fileContent.Routes[0].Paths[i], *path)
		}

		// Verifying that non-string fields remain unchanged
		assert.Equal(t, *fileContent.Services[0].Port, *result.Services[0].Port)
		assert.Equal(t, *fileContent.Services[0].Enabled, *result.Services[0].Enabled)

		// Verifying that references to other entities are preserved
		assert.NotNil(t, result.Routes[0].Service)
		assert.Equal(t, *result.Services[0].Name, *result.Routes[0].Service.Name)
	})

	t.Run("identical input strings result in same sanitized strings", func(t *testing.T) {
		duplicateContent := &file.Content{
			Services: []file.FService{
				{
					Service: kong.Service{
						Name: kong.String("test-service"),
						Host: kong.String("duplicate-value"),
					},
				},
				{
					Service: kong.Service{
						Name: kong.String("another-service"),
						Host: kong.String("duplicate-value"),
					},
				},
			},
		}

		duplicateSanitizer := NewSanitizer(&SanitizerOptions{
			Ctx:     context.Background(),
			Content: duplicateContent,
		})

		duplicateResult, err := duplicateSanitizer.Sanitize()
		require.NoError(t, err)
		require.NotNil(t, duplicateResult)

		require.NotNil(t, duplicateResult.Services)
		require.Len(t, duplicateResult.Services, 2)

		assert.NotNil(t, duplicateResult.Services[0].Host)
		assert.NotNil(t, duplicateResult.Services[1].Host)
		assert.Equal(t, *duplicateResult.Services[0].Host, *duplicateResult.Services[1].Host)
	})

	t.Run("respects entity-level exemptions", func(t *testing.T) {
		content := &file.Content{
			Plugins: []file.FPlugin{
				{
					Plugin: kong.Plugin{
						Name: kong.String("rate-limiting"),
					},
				},
			},
			Routes: []file.FRoute{
				{
					Route: kong.Route{
						Name:    kong.String("test-route"),
						Methods: []*string{kong.String("GET"), kong.String("POST")},
					},
				},
			},
		}

		sanitizer := NewSanitizer(&SanitizerOptions{
			Ctx:     context.Background(),
			Content: content,
		})

		result, err := sanitizer.Sanitize()
		require.NoError(t, err)
		require.NotNil(t, result)

		require.NotNil(t, result.Plugins)
		require.NotNil(t, result.Routes)
		require.Len(t, result.Plugins, 1)
		require.Len(t, result.Routes, 1)

		// Verify that plugin name is preserved
		assert.Equal(t, "rate-limiting", *result.Plugins[0].Name)

		// Verify exempt route methods are preserved
		assert.Equal(t, "GET", *result.Routes[0].Methods[0])
		assert.Equal(t, "POST", *result.Routes[0].Methods[1])
	})

	t.Run("respects config-level exemptions", func(t *testing.T) {
		contentWithID := &file.Content{
			Services: []file.FService{
				{
					Service: kong.Service{
						ID:   kong.String("test-id"),
						Name: kong.String("test-service"),
					},
				},
			},
		}

		sanitizer := NewSanitizer(&SanitizerOptions{
			Ctx:     context.Background(),
			Content: contentWithID.DeepCopy(),
		})

		result, err := sanitizer.Sanitize()
		require.NoError(t, err)
		require.NotNil(t, result)

		require.NotNil(t, result.Services)
		require.Len(t, result.Services, 1)

		assert.Equal(t, *contentWithID.Services[0].ID, *result.Services[0].ID, "ID field should not be sanitized")
	})

	t.Run("sanitizes kong configuration in plugins and partials", func(t *testing.T) {
		fileContent := &file.Content{
			Plugins: []file.FPlugin{
				{
					Plugin: kong.Plugin{
						ID:   kong.String("82c27e99-b1de-4772-aa60-4caa86c0480d"),
						Name: kong.String("rate-limiting-advanced"),
						Config: kong.Configuration{
							"compound_identifier":     nil,
							"consumer_groups":         nil,
							"dictionary_name":         string("kong_rate_limiting_counters"),
							"disable_penalty":         bool(false),
							"enforce_consumer_groups": bool(false),
							"error_code":              float64(429),
							"error_message":           string("API rate limit exceeded"),
							"header_name":             nil,
							"hide_client_headers":     bool(false),
							"identifier":              string("consumer"),
							"limit":                   []any{float64(10)},
							"lock_dictionary_name":    string("kong_locks"),
							"namespace":               string("test-ns"),
							"path":                    nil,
							"retry_after_jitter_max":  float64(0),
							"strategy":                string("local"),
							"sync_rate":               float64(-1),
							"throttling":              nil,
							"window_size":             []any{float64(60)},
							"window_type":             string("fixed"),
						},
						Enabled:   kong.Bool(true),
						Protocols: kong.StringSlice("grpc", "grpcs", "http", "https"),
						Partials: []*kong.PartialLink{
							{
								Partial: &kong.Partial{
									ID: kong.String("13dc230d-d65e-439a-9f05-9fd71abfee4d"),
								},
								Path: kong.String("config.redis"),
							},
						},
					},
				},
			},
			Partials: []file.FPartial{
				{
					Partial: kong.Partial{
						ID:   kong.String("13dc230d-d65e-439a-9f05-9fd71abfee4d"),
						Name: kong.String("redis-ee-common"),
						Type: kong.String("redis-ee"),
						Config: kong.Configuration{
							"cluster_max_redirections": float64(5),
							"cluster_nodes":            nil,
							"connect_timeout":          float64(2000),
							"connection_is_proxied":    bool(false),
							"database":                 float64(0),
							"host":                     string("127.0.0.1"),
							"keepalive_backlog":        nil,
							"keepalive_pool_size":      float64(256),
							"password":                 nil,
							"port":                     float64(6379),
							"read_timeout":             float64(3001),
							"send_timeout":             float64(2004),
							"sentinel_master":          nil,
							"sentinel_nodes":           nil,
							"sentinel_password":        nil,
							"sentinel_role":            nil,
							"sentinel_username":        nil,
							"server_name":              nil,
							"ssl":                      bool(false),
							"ssl_verify":               bool(false),
							"username":                 nil,
						},
						Tags: kong.StringSlice("redis-partials"),
					},
				},
			},
		}

		sanitizer := NewSanitizer(&SanitizerOptions{
			Ctx:     context.Background(),
			Content: fileContent.DeepCopy(),
		})

		result, err := sanitizer.Sanitize()
		require.NoError(t, err)
		require.NotNil(t, result)

		require.NotNil(t, result.Plugins)
		require.NotNil(t, result.Partials)
		require.Len(t, result.Plugins, 1)
		require.Len(t, result.Partials, 1)

		originalPluginConfigMap := map[string]interface{}(fileContent.Plugins[0].Config)
		sanitizedPluginConfigMap := map[string]interface{}(result.Plugins[0].Config)
		verifyConfigSanitization(t, originalPluginConfigMap, sanitizedPluginConfigMap)

		originalPartialConfigMap := map[string]interface{}(fileContent.Partials[0].Config)
		sanitizedPartialConfigMap := map[string]interface{}(result.Partials[0].Config)
		verifyConfigSanitization(t, originalPartialConfigMap, sanitizedPartialConfigMap)
	})
}

func Test_sanitizeConfig(t *testing.T) {
	tests := []struct {
		name         string
		config       interface{}
		shouldExempt bool
	}{
		{
			name: "simple kong configuration",
			config: map[string]interface{}{
				"request_jq_program": "select(.name == \"James Dean\").name = \"John Doe\"\n",
				"request_if_media_type": []interface{}{
					"application/json",
				},
				"request_jq_program_options": map[string]interface{}{
					"ascii_output":   false,
					"compact_output": true,
					"join_output":    false,
					"raw_output":     false,
					"sort_keys":      false,
				},
			},
		},
		{
			name: "kong configuration with multiple nested levels",
			config: map[string]interface{}{
				"behavior": map[string]interface{}{
					"idp_error_response_body_template":  "{ \"code\": \"{{status}}\", \"message\": \"{{message}}\" }",
					"idp_error_response_content_type":   "application/json; charset=utf-8",
					"upstream_access_token_header_name": "Authorization",
				},
				"oauth": map[string]interface{}{
					"client_id":      "clientOne",
					"client_secret":  nil,
					"password":       "my-password",
					"scopes":         []interface{}{"openid"},
					"token_endpoint": "http://example.com",
					"username":       "userOne",
				},
				"cache": map[string]interface{}{
					"strategy": "memory",
					"redis": map[string]interface{}{
						"host":        "127.0.0.1",
						"port":        6379,
						"password":    "secret",
						"server_name": nil,
					},
				},
			},
		},
		{
			name:   "direct string input",
			config: "sensitive-data",
		},
		{
			name:         "empty string input",
			config:       "",
			shouldExempt: true,
		},
		{
			name:         "number input: float",
			config:       float64(429.0),
			shouldExempt: true,
		},
		{
			name:         "number input: int",
			config:       int64(429),
			shouldExempt: true,
		},
		{
			name:         "boolean input",
			config:       bool(true),
			shouldExempt: true,
		},
		{
			name:         "null input",
			config:       nil,
			shouldExempt: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sanitizer := NewSanitizer(&SanitizerOptions{
				Ctx: context.Background(),
			})
			sanitizedConfig := sanitizer.sanitizeConfig(reflect.ValueOf(tc.config))

			if tc.shouldExempt {
				assert.Equal(t, tc.config, sanitizedConfig, "Exempted field should not be sanitized")
				return
			}

			assert.NotEqual(t, sanitizedConfig, tc.config)
			verifyConfigSanitization(t, tc.config, sanitizedConfig)
		})
	}
}

// Helper function to recursively check that config was sanitized
func verifyConfigSanitization(t *testing.T, original, sanitized interface{}) {
	t.Helper()

	switch orig := original.(type) {
	case map[string]interface{}:
		sanitizedMap, ok := sanitized.(map[string]interface{})
		require.True(t, ok, "Sanitized value should be a map")

		for k, v := range orig {
			if shouldSkipSanitization("", k, nil) {
				assert.Equal(t, v, sanitizedMap[k], "Exempted field %s should not be sanitized", k)
				continue
			}

			// Recursively check nested structures
			if sanitizedVal, ok := sanitizedMap[k]; ok {
				verifyConfigSanitization(t, v, sanitizedVal)
			} else {
				t.Errorf("Key %s missing from sanitized map", k)
			}
		}
	case []interface{}:
		sanitizedSlice, ok := sanitized.([]interface{})
		require.True(t, ok, "Sanitized value should be a slice")
		require.Equal(t, len(orig), len(sanitizedSlice), "Slice length should be preserved")

		for i, v := range orig {
			verifyConfigSanitization(t, v, sanitizedSlice[i])
		}
	case string:
		sanitizedStr, ok := sanitized.(string)
		require.True(t, ok, "Sanitized value should be a string")
		assert.NotEqual(t, orig, sanitizedStr, "String value should be sanitized: %s", orig)
	default:
		assert.Equal(t, orig, sanitized, "Non-string primitives should remain unchanged")
	}
}
