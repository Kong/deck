package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyncCmd_ZeroParallelism(t *testing.T) {
	cmd := newSyncCmd(false)
	syncCmdParallelism = 0
	err := cmd.PreRunE(cmd, []string{"-"})
	require.Error(t, err)
	require.EqualError(t, err, "--parallelism cannot be less than 1, got 0")
}

func TestSyncCmd_ValidParallelism(t *testing.T) {
	cmd := newSyncCmd(false)
	syncCmdParallelism = 10
	err := cmd.PreRunE(cmd, []string{"-"})
	require.NoError(t, err)
}

func TestSyncCmd_Deprecated_NegativeParallelism(t *testing.T) {
	cmd := newSyncCmd(true)
	syncCmdParallelism = -1
	err := cmd.PreRunE(cmd, nil)
	require.Error(t, err)
	require.EqualError(t, err, "--parallelism cannot be less than 1, got -1")
}

func TestSyncCmd_Deprecated_ValidParallelism(t *testing.T) {
	cmd := newSyncCmd(true)
	syncCmdParallelism = 10
	err := cmd.PreRunE(cmd, nil)
	require.NoError(t, err)
}
