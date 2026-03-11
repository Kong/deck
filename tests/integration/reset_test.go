//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var caCert = &kong.CACertificate{
	CertDigest: kong.String("34e0f1f3d83faefcc8514b6295bc822eab1110dc120140ddf342c017baee8c0f"),
	Cert: kong.String(`-----BEGIN CERTIFICATE-----
MIIDkzCCAnugAwIBAgIUYGc07pbHSjOBPreXh7OcNT2+sD4wDQYJKoZIhvcNAQEL
BQAwWTELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIs
IEluYy4xJjAkBgNVBAMMHVlvbG80MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMB4X
DTIyMDMyOTE5NDczM1oXDTMyMDMyNjE5NDczM1owWTELMAkGA1UEBhMCVVMxCzAJ
BgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIsIEluYy4xJjAkBgNVBAMMHVlvbG80
MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAvnhTgdJALnuLKDA0ZUZRVMqcaaC+qvfJkiEFGYwX2ZJiFtzU65F/
sB2L0ToFqY4tmMVlOmiSZFnRLDZecmQDbbNwc3wtNikmxIOzx4qR4kbRP8DDdyIf
gaNmGCuaXTM5+FYy2iNBn6CeibIjqdErQlAbFLwQs5t3mLsjii2U4cyvfRtO+0RV
HdJ6Np5LsVziN0c5gVIesIrrbxLcOjtXDzwd/w/j5NXqL/OwD5EBH2vqd3QKKX4t
s83BLl2EsbUse47VAImavrwDhmV6S/p/NuJHqjJ6dIbXLYxNS7g26ijcrXxvNhiu
YoZTykSgdI3BXMNAm1ahP/BtJPZpU7CVdQIDAQABo1MwUTAdBgNVHQ4EFgQUe1WZ
fMfZQ9QIJIttwTmcrnl40ccwHwYDVR0jBBgwFoAUe1WZfMfZQ9QIJIttwTmcrnl4
0ccwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAs4Z8VYbvEs93
haTHdbbaKk0V6xAL/Q8I8GitK9E8cgf8C5rwwn+wU/Gf39dtMUlnW8uxyzRPx53u
CAAcJAWkabT+xwrlrqjO68H3MgIAwgWA5yZC+qW7ECA8xYEK6DzEHIaOpagJdKcL
IaZr/qTJlEQClvwDs4x/BpHRB5XbmJs86GqEB7XWAm+T2L8DluHAXvek+welF4Xo
fQtLlNS/vqTDqPxkSbJhFv1L7/4gdwfAz51wH/iL7AG/ubFEtoGZPK9YCJ40yTWz
8XrUoqUC+2WIZdtmo6dFFJcLfQg4ARJZjaK6lmxJun3iRMZjKJdQKm/NEKz4y9kA
u8S6yNlu2Q==
-----END CERTIFICATE-----`),
}

func Test_Reset_SkipCACert_2x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "reset with --skip-ca-certificates should ignore CA certs",
			kongFile: "testdata/reset/001-skip-ca-cert/kong.yaml",
			expectedState: utils.KongRawState{
				CACertificates: []*kong.CACertificate{caCert},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// ca_certificates first appeared in 1.3, but we limit to 2.7+
			// here because the schema changed and the entities aren't the same
			// across all versions, even though the skip functionality works the same.
			runWhen(t, "kong", ">=2.7.0 <3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			reset(t, "--skip-ca-certificates")
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

func Test_Reset_SkipCACert_3x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "reset with --skip-ca-certificates should ignore CA certs",
			kongFile: "testdata/reset/001-skip-ca-cert/kong3x.yaml",
			expectedState: utils.KongRawState{
				CACertificates: []*kong.CACertificate{caCert},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// ca_certificates first appeared in 1.3, but we limit to 2.7+
			// here because the schema changed and the entities aren't the same
			// across all versions, even though the skip functionality works the same.
			runWhen(t, "kong", ">=3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			reset(t, "--skip-ca-certificates")
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

func Test_Reset_ConsumerGroupConsumersWithCustomID(t *testing.T) {
	runWhen(t, "enterprise", ">=3.0.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	require.NoError(t, sync(context.Background(), "testdata/sync/028-consumer-group-consumers-custom_id/kong.yaml"))
	reset(t)
	testKongState(t, client, false, false, utils.KongRawState{}, nil)
}

func Test_Reset_ConsumerGroupConsumersWithCustomID_Konnect(t *testing.T) {
	runWhenKonnect(t)
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	require.NoError(t, sync(context.Background(), "testdata/sync/028-consumer-group-consumers-custom_id/kong.yaml"))
	reset(t)
	testKongState(t, client, true, false, utils.KongRawState{}, nil)
}

func Test_Reset_KonnectWorkspace(t *testing.T) {
	runWhenKonnect(t)
	setup(t)

	ctx := context.Background()

	tests := []struct {
		name        string
		stateFile   string
		resetFlags  []string
		notContains string // if set, verify dump output does NOT contain this after reset
	}{
		{
			name:        "reset workspace with single route",
			stateFile:   "testdata/reset/002-konnect-workspace/single-route.yaml",
			resetFlags:  []string{"--workspace", "test-workspace"},
			notContains: "route-reset-1",
		},
		{
			name:        "reset workspace with multiple routes",
			stateFile:   "testdata/reset/002-konnect-workspace/multiple-routes.yaml",
			resetFlags:  []string{"--workspace", "test-workspace"},
			notContains: "route-reset",
		},
		{
			name:        "reset without workspace flag (CP-level)",
			stateFile:   "testdata/reset/002-konnect-workspace/no-workspace.yaml",
			resetFlags:  []string{},
			notContains: "route-no-workspace",
		},
		{
			name:        "reset with default workspace name",
			stateFile:   "testdata/reset/002-konnect-workspace/default-workspace.yaml",
			resetFlags:  []string{"--workspace", "default"},
			notContains: "route-default-ws",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			require.NoError(t, sync(ctx, tc.stateFile))

			// Verify entity exists before reset
			dumpFlags := []string{"-o", "-"}
			if len(tc.resetFlags) > 0 {
				dumpFlags = append(dumpFlags, tc.resetFlags...)
			}
			output, err := dump(dumpFlags...)
			require.NoError(t, err)
			assert.Contains(t, output, tc.notContains,
				"Entity should exist before reset")

			// Reset with the specified flags
			reset(t, tc.resetFlags...)

			// Verify entity is gone after reset
			output, err = dump(dumpFlags...)
			require.NoError(t, err)
			assert.NotContains(t, output, tc.notContains,
				"Entity should NOT exist after reset")
		})
	}
}

func Test_Reset_KonnectWorkspace_Isolation(t *testing.T) {
	runWhenKonnect(t)
	setup(t)

	ctx := context.Background()

	// Reset everything first
	reset(t)

	// Sync routes to two different workspaces
	require.NoError(t, sync(ctx, "testdata/reset/002-konnect-workspace/workspace1-route.yaml"))
	require.NoError(t, sync(ctx, "testdata/reset/002-konnect-workspace/workspace2-route.yaml"))

	// Verify both routes exist
	output1, err := dump("-o", "-", "--workspace", "workspace1")
	require.NoError(t, err)
	assert.Contains(t, output1, "route-workspace1", "workspace1 route should exist")

	output2, err := dump("-o", "-", "--workspace", "workspace2")
	require.NoError(t, err)
	assert.Contains(t, output2, "route-workspace2", "workspace2 route should exist")

	// Reset only workspace1
	reset(t, "--workspace", "workspace1")

	// Verify workspace1 is empty
	output1, err = dump("-o", "-", "--workspace", "workspace1")
	require.NoError(t, err)
	assert.NotContains(t, output1, "route-workspace1",
		"workspace1 route should be deleted after reset")

	// Verify workspace2 still has its route (isolation check)
	output2, err = dump("-o", "-", "--workspace", "workspace2")
	require.NoError(t, err)
	assert.Contains(t, output2, "route-workspace2",
		"workspace2 route should NOT be affected by workspace1 reset")
}

func Test_Reset_KonnectWorkspace_AllWorkspaces(t *testing.T) {
	runWhenKonnect(t)
	setup(t)

	ctx := context.Background()
	reset(t)

	// Sync routes to two different workspaces
	require.NoError(t, sync(ctx, "testdata/reset/002-konnect-workspace/workspace1-route.yaml"))
	require.NoError(t, sync(ctx, "testdata/reset/002-konnect-workspace/workspace2-route.yaml"))

	// Verify both routes exist
	output1, err := dump("-o", "-", "--workspace", "workspace1")
	require.NoError(t, err)
	assert.Contains(t, output1, "route-workspace1", "workspace1 route should exist")

	output2, err := dump("-o", "-", "--workspace", "workspace2")
	require.NoError(t, err)
	assert.Contains(t, output2, "route-workspace2", "workspace2 route should exist")

	// Reset all workspaces
	reset(t, "--all-workspaces")

	// Verify workspace1 is empty
	output1, err = dump("-o", "-", "--workspace", "workspace1")
	require.NoError(t, err)
	assert.NotContains(t, output1, "route-workspace1",
		"workspace1 route should be deleted after --all-workspaces reset")

	// Verify workspace2 is also empty
	output2, err = dump("-o", "-", "--workspace", "workspace2")
	require.NoError(t, err)
	assert.NotContains(t, output2, "route-workspace2",
		"workspace2 route should be deleted after --all-workspaces reset")
}
