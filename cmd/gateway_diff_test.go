package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiffCmd_ZeroParallelism(t *testing.T) {
	cmd := newDiffCmd(false)
	diffCmdParallelism = 0
	err := cmd.PreRunE(cmd, []string{"-"})
	require.Error(t, err)
	require.EqualError(t, err, "--parallelism cannot be less than 1, got 0")
}

func TestDiffCmd_ValidParallelism(t *testing.T) {
	cmd := newDiffCmd(false)
	diffCmdParallelism = 10
	err := cmd.PreRunE(cmd, []string{"-"})
	require.NoError(t, err)
}

func TestDiffCmd_Deprecated_NegativeParallelism(t *testing.T) {
	cmd := newDiffCmd(true)
	diffCmdParallelism = -1
	diffCmdKongStateFile = []string{"kong.yaml"}
	err := cmd.PreRunE(cmd, nil)
	require.Error(t, err)
	require.EqualError(t, err, "--parallelism cannot be less than 1, got -1")
}

func TestDiffCmd_Deprecated_ValidParallelism(t *testing.T) {
	cmd := newDiffCmd(true)
	diffCmdParallelism = 10
	diffCmdKongStateFile = []string{"kong.yaml"}
	err := cmd.PreRunE(cmd, nil)
	require.NoError(t, err)
}
