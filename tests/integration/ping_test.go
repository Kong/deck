//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_KonnectPing(t *testing.T) {
	t.Run("konnect ping", func(t *testing.T) {
		runWhen(t, "konnect", "")
		require.NoError(t, ping())
	})
}
