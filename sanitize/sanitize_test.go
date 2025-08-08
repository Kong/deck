package sanitize

import (
	"context"
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
}
