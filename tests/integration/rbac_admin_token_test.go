//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_RBAC_AdminToken exercises the --kong-admin-token flag (added to
// authenticate against an RBAC-enabled Kong Admin API).
//
// It only runs against an Enterprise Kong instance with RBAC enforcement
// enabled, which is provided by the dedicated `integration-rbac` job in
// .github/workflows/integration-enterprise.yaml. When RBAC is enforced, a
// command that talks to the Admin API must fail without a valid admin token
// and succeed once the token is supplied.
func Test_RBAC_AdminToken(t *testing.T) {
	runWhenRBAC(t, ">=2.8.0")

	// disable analytics for integration tests
	t.Setenv("DECK_ANALYTICS", "off")

	// The CI job seeds the kong_admin token in KONG_ADMIN_TOKEN; fall back to
	// the decK CLI variable for local runs.
	adminToken := os.Getenv("KONG_ADMIN_TOKEN")
	if adminToken == "" {
		adminToken = os.Getenv("DECK_KONG_ADMIN_TOKEN")
	}
	require.NotEmpty(t, adminToken,
		"KONG_ADMIN_TOKEN or DECK_KONG_ADMIN_TOKEN must be set when running RBAC tests")

	// online validation hits the Admin API but does not mutate state, so no
	// reset/cleanup (which would itself require the token) is needed.
	const stateFile = "testdata/validate/kong-ee.yaml"

	t.Run("fails when kong-admin-token is not passed", func(t *testing.T) {
		// make sure the CLI cannot pick the token up from the environment, so
		// the request reaches Kong unauthenticated.
		t.Setenv("DECK_KONG_ADMIN_TOKEN", "")

		err := validate(ONLINE, stateFile)
		require.Error(t, err,
			"online validate should fail against an RBAC-enabled Kong without an admin token")
		// Assert it fails *specifically* because of authentication (HTTP 401)
		// rather than some unrelated error (bad file, gateway down, etc.).
		// go-kong formats API errors as `HTTP status 401 (message: ...)`.
		assert.Contains(t, err.Error(), "401",
			"expected an authentication failure (HTTP 401), got: %v", err)
	})

	t.Run("succeeds when kong-admin-token is passed", func(t *testing.T) {
		// scrub the env so the token can only come from the flag, proving the
		// flag is what authenticates the request.
		t.Setenv("DECK_KONG_ADMIN_TOKEN", "")

		err := validate(ONLINE, stateFile, "--kong-admin-token", adminToken)
		require.NoError(t, err,
			"online validate should succeed against an RBAC-enabled Kong with a valid admin token")
	})
}
