package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{
			name:           "render with traditional route",
			stateFile:      "testdata/render/003-traditional-routes/input.yaml",
			additionalArgs: []string{},
			expectedFile:   "testdata/render/003-traditional-routes/expected.yaml",
			envVars:        map[string]string{},
		},
		{
			name:           "render with expression route",
			stateFile:      "testdata/render/004-expression-routes/input.yaml",
			additionalArgs: []string{},
			expectedFile:   "testdata/render/004-expression-routes/expected.yaml",
			envVars:        map[string]string{},
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
			require.NoError(t, err)

			expected, err := readFile(tc.expectedFile)
			require.NoError(t, err)
			assert.Equal(t, expected, output)
		})
	}
}
