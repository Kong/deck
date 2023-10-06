//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_KonnectPing(t *testing.T) {
	t.Run("konnect ping - email/password", func(t *testing.T) {
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
