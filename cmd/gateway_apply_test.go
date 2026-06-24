package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyCmd_ZeroParallelism(t *testing.T) {
	cmd := newApplyCmd()
	applyCmdParallelism = 0
	err := cmd.PreRunE(cmd, []string{"-"})
	require.Error(t, err)
	require.EqualError(t, err, "--parallelism cannot be less than 1, got 0")
}

func TestApplyCmd_NegativeParallelism(t *testing.T) {
	cmd := newApplyCmd()
	applyCmdParallelism = -5
	err := cmd.PreRunE(cmd, []string{"-"})
	require.Error(t, err)
	require.EqualError(t, err, "--parallelism cannot be less than 1, got -5")
}

func TestApplyCmd_ValidParallelism(t *testing.T) {
	cmd := newApplyCmd()
	applyCmdParallelism = 10
	err := cmd.PreRunE(cmd, []string{"-"})
	require.NoError(t, err)
}
