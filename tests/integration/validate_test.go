//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Validate_Konnect(t *testing.T) {
	setup(t)
	runWhen(t, "konnect", "")

	tests := []struct {
		name           string
		stateFile      string
		additionalArgs []string
		errorExpected  bool
	}{
		{
			name:           "validate with konnect",
			stateFile:      "testdata/validate/konnect.yaml",
			additionalArgs: []string{},
		},
		{
			name:           "validate with --konnect-compatibility",
			stateFile:      "testdata/validate/konnect.yaml",
			additionalArgs: []string{"--konnect-compatibility"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validateOpts := []string{
				tc.stateFile,
			}
			validateOpts = append(validateOpts, tc.additionalArgs...)

			err := validate(true, validateOpts...)
			assert.NoError(t, err)
		})
	}
}

func Test_Validate_Konnect_RBAC(t *testing.T) {
	setup(t)
	runWhen(t, "konnect", "")

	tests := []struct {
		name           string
		stateFile      string
		additionalArgs []string
		errorExpected  bool
	}{
		{
			name:           "validate with --rbac-resources-only",
			stateFile:      "testdata/validate/rbac-resources.yaml",
			additionalArgs: []string{"--rbac-resources-only"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validateOpts := []string{
				tc.stateFile,
			}
			validateOpts = append(validateOpts, tc.additionalArgs...)

			err := validate(true, validateOpts...)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "rbac validation not yet supported in konnect mode")
		})
	}
}
