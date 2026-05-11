//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SyncDiagnosticSeverityOverrides_RLAWarningAsError(t *testing.T) {
	runWhen(t, "enterprise", ">=3.4.0")
	setup(t)

	err := sync(
		context.Background(),
		"testdata/render/005-diagnostics/rla-consumer-groups.yaml",
		"-E", "rla-consumer-groups-deprecated",
	)
	require.Error(t, err)
	assert.ErrorContains(t, err, errorConsumerGroupPolicies)
}
