package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RenderDiagnosticSeverityOverrides_RouteRegexWarningAsError(t *testing.T) {
	skipWhenKonnect(t)
	_, err := render(
		"testdata/render/005-diagnostics/route-regex.yaml",
		"-E", "route-regex-path-format",
	)
	require.Error(t, err)
	require.ErrorContains(t, err, "warning (route-regex-path-format)")
}

func Test_RenderDiagnosticSeverityOverrides_OIDCSeverityOverrides(t *testing.T) {
	skipWhenKonnect(t)
	t.Run("defaults to hard error", func(t *testing.T) {
		_, err := render("testdata/render/005-diagnostics/oidc-missing-config.yaml")
		require.Error(t, err)
		require.ErrorContains(t, err, "cache_tokens_salt")
	})

	t.Run("downgrades to warning", func(t *testing.T) {
		output, err := render(
			"testdata/render/005-diagnostics/oidc-missing-config.yaml",
			"-W", "oidc-missing-required-config",
		)
		require.NoError(t, err)
		assert.Contains(t, output, "openid-connect")
	})

	t.Run("error wins on conflict", func(t *testing.T) {
		_, err := render(
			"testdata/render/005-diagnostics/oidc-missing-config.yaml",
			"-E", "oidc-missing-required-config",
			"-W", "oidc-missing-required-config",
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "cache_tokens_salt")
	})
}

func Test_RenderDiagnosticSeverityOverrides_DefaultWarningsRemainWarnings(t *testing.T) {
	skipWhenKonnect(t)
	output, err := render("testdata/render/005-diagnostics/route-regex.yaml")
	require.NoError(t, err)
	assert.Contains(t, output, "svc-route-regex")
}

func Test_RenderDiagnosticSeverityOverrides_InvalidCode(t *testing.T) {
	skipWhenKonnect(t)
	_, err := render(
		"testdata/render/005-diagnostics/route-regex.yaml",
		"-E", "not-a-real-code",
	)
	require.Error(t, err)
	require.ErrorContains(t, err, "Valid diagnostic codes")
	require.ErrorContains(t, err, "route-regex-path-format")
	require.ErrorContains(t, err, "rla-consumer-groups-deprecated")
	require.ErrorContains(t, err, "oidc-missing-required-config")
}
