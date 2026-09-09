//go:build integration

package integration

import (
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/require"
)

// Test_IncrementalSync_Smoke skips unless the test Kong Gateway was actually started
// with KONG_INCREMENTAL_SYNC=on (see .ci/setup_kong_ee.sh) and confirms the setting
// reached the Gateway container by reading it back from the Admin API root endpoint.
// Incremental config sync is Enterprise-only, hybrid-mode-only, GA from 3.10 onward.
func Test_IncrementalSync_Smoke(t *testing.T) {
	runWhen(t, "enterprise", ">=3.10.0")
	kong.RunWhenKongConfigEnabled(t, "incremental_sync")

	client, err := getTestClient()
	require.NoError(t, err)

	info, err := client.Root(t.Context())
	require.NoError(t, err)

	configuration, ok := info["configuration"].(map[string]interface{})
	require.True(t, ok, "expected 'configuration' in Admin API root response")
	require.Equal(t, "on", configuration["incremental_sync"], "KONG_INCREMENTAL_SYNC did not reach the Gateway container")
}
