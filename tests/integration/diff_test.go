//go:build integration

package integration

import (
	"testing"

	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

// test scope:
//   - 1.x
//   - 2.x
func Test_Diff_Workspace_OlderThan3x(t *testing.T) {
	tests := []struct {
		name          string
		stateFile     string
		expectedState utils.KongRawState
	}{
		{
			name:      "diff with not existent workspace doesn't error out",
			stateFile: "testdata/diff/001-not-existing-workspace/kong.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, "<3.0.0")
			teardown := setup(t)
			defer teardown(t)

			err := diff(tc.stateFile)
			assert.Nil(t, err)
		})
	}
}

// test scope:
//   - 3.x
func Test_Diff_Workspace_NewerThan3x(t *testing.T) {
	tests := []struct {
		name          string
		stateFile     string
		expectedState utils.KongRawState
	}{
		{
			name:      "diff with not existent workspace doesn't error out",
			stateFile: "testdata/diff/001-not-existing-workspace/kong3x.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kong.RunWhenKong(t, ">=3.0.0")
			teardown := setup(t)
			defer teardown(t)

			err := diff(tc.stateFile)
			assert.Nil(t, err)
		})
	}
}
