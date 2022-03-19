//go:build integration

package integration

import (
	"testing"

	"github.com/kong/deck/utils"
	"github.com/stretchr/testify/assert"
)

func Test_Diff_Workspace(t *testing.T) {
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
			teardown := setup(t)
			defer teardown(t)

			err := diff(tc.stateFile)
			assert.Nil(t, err)
		})
	}
}
