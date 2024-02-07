package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RenderPlain(t *testing.T) {
	tests := []struct {
		name           string
		stateFile      string
		additionalArgs []string
		expectedFile   string
		envVars        map[string]string
	}{
		{
			name:           "render with mocked env",
			stateFile:      "testdata/render/001-mocked-env/input.yaml",
			additionalArgs: []string{},
			expectedFile:   "testdata/render/001-mocked-env/expected.yaml",
			envVars:        map[string]string{},
		},
		{
			name:           "render with populated env",
			stateFile:      "testdata/render/002-populated-env/input.yaml",
			additionalArgs: []string{"--populate-env-vars"},
			expectedFile:   "testdata/render/002-populated-env/expected.yaml",
			envVars: map[string]string{
				"DECK_MOCKBIN_HOST":    "mockbin.org",
				"DECK_MOCKBIN_ENABLED": "true",
				"DECK_WRITE_TIMEOUT":   "777",
				"DECK_FOO_FLOAT":       "123",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			renderOpts := []string{
				tc.stateFile,
			}
			renderOpts = append(renderOpts, tc.additionalArgs...)

			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}

			output, err := render(renderOpts...)
			assert.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			assert.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}
