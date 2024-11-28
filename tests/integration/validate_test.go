//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	ONLINE  = true
	OFFLINE = false
)

func Test_Validate_Konnect(t *testing.T) {
	setup(t)
	runWhen(t, "konnect", "")

	tests := []struct {
		name           string
		stateFile      string
		additionalArgs []string
		errorExpected  bool
		errorString    string
	}{
		{
			name:           "validate with konnect",
			stateFile:      "testdata/validate/konnect.yaml",
			additionalArgs: []string{},
			errorExpected:  false,
		},
		{
			name:           "validate with --konnect-compatibility",
			stateFile:      "testdata/validate/konnect.yaml",
			additionalArgs: []string{"--konnect-compatibility"},
			errorExpected:  false,
		},
		{
			name:           "validate with 1.1 version file",
			stateFile:      "testdata/validate/konnect_1_1.yaml",
			additionalArgs: []string{},
			errorExpected:  true,
			errorString:    "[version] decK file version must be '3.0' or greater",
		},
		{
			name:           "validate with no version in deck file",
			stateFile:      "testdata/validate/konnect_no_version.yaml",
			additionalArgs: []string{},
			errorExpected:  true,
			errorString:    "[version] unable to determine decK file version",
		},
		{
			name:           "validate with --rbac-resources-only",
			stateFile:      "testdata/validate/rbac-resources.yaml",
			additionalArgs: []string{"--rbac-resources-only"},
			errorExpected:  true,
			errorString:    "[rbac] not yet supported by konnect",
		},
		{
			name:           "validate with workspace set",
			stateFile:      "testdata/validate/konnect.yaml",
			additionalArgs: []string{"--workspace=default"},
			errorExpected:  true,
			errorString:    "[workspaces] not supported by Konnect - use control planes instead",
		},
		{
			name:           "validate with no konnect config in file, passed via cli flag konnect control plane",
			stateFile:      "testdata/validate/konnect_invalid.yaml",
			additionalArgs: []string{"--konnect-control-plane-name=default"},
			errorExpected:  false,
		},
		{
			name:           "validate with no konnect config in file, passed via cli flag konnect runtime group",
			stateFile:      "testdata/validate/konnect_invalid.yaml",
			additionalArgs: []string{"--konnect-runtime-group-name=default"},
			errorExpected:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validateOpts := append([]string{
				tc.stateFile,
			}, tc.additionalArgs...)

			err := validate(ONLINE, validateOpts...)

			if tc.errorExpected {
				assert.Error(t, err)
				if tc.errorString != "" {
					assert.Contains(t, err.Error(), tc.errorString)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

func Test_Validate_File(t *testing.T) {
	setup(t)

	tests := []struct {
		name           string
		stateFile      string
		additionalArgs []string
		errorExpected  bool
	}{
		{
			name:           "file validate format version 1.1",
			stateFile:      "testdata/validate/kong.yaml",
			additionalArgs: []string{},
		},
		{
			name:           "file validate format version 3.0",
			stateFile:      "testdata/validate/kong3x.yaml",
			additionalArgs: []string{},
		},
		{
			name:           "file validate with --konnect-compatibility",
			stateFile:      "testdata/validate/konnect.yaml",
			additionalArgs: []string{"--konnect-compatibility"},
		},
		{
			name:           "file validate with --workspace",
			stateFile:      "testdata/validate/kong3x.yaml",
			additionalArgs: []string{"--workspace=default"},
		},
		{
			name:           "file validate with --rbac-resources-only",
			stateFile:      "testdata/validate/rbac-resources.yaml",
			additionalArgs: []string{"--rbac-resources-only"},
		},
		{
			name:           "file validate with --check-online-plugins-only",
			stateFile:      "testdata/validate/kong3x.yaml",
			additionalArgs: []string{"--check-online-plugins-only"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validateOpts := []string{
				tc.stateFile,
			}
			validateOpts = append(validateOpts, tc.additionalArgs...)

			err := validate(OFFLINE, validateOpts...)
			assert.NoError(t, err)
		})
	}
}

func Test_Validate_Gateway(t *testing.T) {
	setup(t)
	runWhen(t, "kong", ">=2.8.0")

	tests := []struct {
		name           string
		stateFile      string
		additionalArgs []string
		errorExpected  bool
	}{
		{
			name:           "validate format version 1.1",
			stateFile:      "testdata/validate/kong.yaml",
			additionalArgs: []string{},
		},
		{
			name:           "validate format version 3.0",
			stateFile:      "testdata/validate/kong3x.yaml",
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

			err := validate(ONLINE, validateOpts...)
			assert.NoError(t, err)
		})
	}
}

func Test_Validate_Gateway_EE(t *testing.T) {
	setup(t)
	runWhen(t, "enterprise", ">=2.8.0")

	tests := []struct {
		name           string
		stateFile      string
		additionalArgs []string
		errorExpected  bool
	}{
		{
			name:           "validate format version 1.1",
			stateFile:      "testdata/validate/kong.yaml",
			additionalArgs: []string{},
		},
		{
			name:           "validate format version 3.0",
			stateFile:      "testdata/validate/kong-ee.yaml",
			additionalArgs: []string{},
		},
		{
			name:           "validate with --konnect-compatibility",
			stateFile:      "testdata/validate/konnect.yaml",
			additionalArgs: []string{"--konnect-compatibility"},
		},
		{
			name:           "validate with --workspace",
			stateFile:      "testdata/validate/kong-ee.yaml",
			additionalArgs: []string{"--workspace=default"},
		},
		// TODO: Add a rbac flag test, once the behaviour is fixed
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validateOpts := []string{
				tc.stateFile,
			}
			validateOpts = append(validateOpts, tc.additionalArgs...)

			err := validate(ONLINE, validateOpts...)
			assert.NoError(t, err)
		})
	}
}
