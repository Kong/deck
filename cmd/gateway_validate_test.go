package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateCmd_InvalidParallelism(t *testing.T) {
	cmd := newValidateCmd(false, true)
	validateParallelism = 0
	err := cmd.PreRunE(cmd, []string{"-"})
	require.Error(t, err)
	require.EqualError(t, err, "--parallelism cannot be less than 1, got 0")
}

func TestValidateCmd_ValidParallelism(t *testing.T) {
	cmd := newValidateCmd(false, true)
	validateParallelism = 10
	err := cmd.PreRunE(cmd, []string{"-"})
	require.NoError(t, err)
}
