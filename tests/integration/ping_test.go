//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_KonnectPing(t *testing.T) {
	t.Run("konnect ping - email/password", func(t *testing.T) {
		// TODO: https://github.com/Kong/deck/issues/1100
		// Remove the test altogether or fix the basic auth support.
		t.Skip("https://github.com/Kong/deck/issues/1100")

		runWhen(t, "konnect", "")
		require.NoError(t, ping(
			"--konnect-email", os.Getenv("DECK_KONNECT_EMAIL"),
			"--konnect-password", os.Getenv("DECK_KONNECT_PASSWORD"),
		))
	})

	t.Run("konnect ping - token", func(t *testing.T) {
		runWhen(t, "konnect", "")
		require.NoError(t, ping(
			"--konnect-token", os.Getenv("DECK_KONNECT_TOKEN"),
		))
	})
}
