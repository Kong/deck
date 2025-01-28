//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Apply_3x(t *testing.T) {
	// setup stage

	tests := []struct {
		name          string
		firstFile     string
		secondFile    string
		expectedState string
		runWhen       string
	}{
		{
			name:          "applies multiple of the same entity",
			firstFile:     "testdata/apply/001-same-type/service-01.yaml",
			secondFile:    "testdata/apply/001-same-type/service-02.yaml",
			expectedState: "testdata/apply/001-same-type/expected-state.yaml",
			runWhen:       "kong",
		},
		{
			name:          "applies different entity types",
			firstFile:     "testdata/apply/002-different-types/service-01.yaml",
			secondFile:    "testdata/apply/002-different-types/plugin-01.yaml",
			expectedState: "testdata/apply/002-different-types/expected-state.yaml",
			runWhen:       "kong",
		},
		{
			name:          "accepts consumer foreign keys",
			firstFile:     "testdata/apply/003-foreign-keys-consumers/consumer-01.yaml",
			secondFile:    "testdata/apply/003-foreign-keys-consumers/plugin-01.yaml",
			expectedState: "testdata/apply/003-foreign-keys-consumers/expected-state.yaml",
			runWhen:       "kong",
		},
		{
			name:          "accepts consumer group foreign keys",
			firstFile:     "testdata/apply/004-foreign-keys-consumer-groups/consumer-group-01.yaml",
			secondFile:    "testdata/apply/004-foreign-keys-consumer-groups/consumer-01.yaml",
			expectedState: "testdata/apply/004-foreign-keys-consumer-groups/expected-state.yaml",
			runWhen:       "enterprise",
		},
		{
			name:          "accepts service foreign keys",
			firstFile:     "testdata/apply/005-foreign-keys-services/service-01.yaml",
			secondFile:    "testdata/apply/005-foreign-keys-services/plugin-01.yaml",
			expectedState: "testdata/apply/005-foreign-keys-services/expected-state.yaml",
			runWhen:       "kong",
		},
		{
			name:          "accepts route foreign keys",
			firstFile:     "testdata/apply/006-foreign-keys-routes/route-01.yaml",
			secondFile:    "testdata/apply/006-foreign-keys-routes/plugin-01.yaml",
			expectedState: "testdata/apply/006-foreign-keys-routes/expected-state.yaml",
			runWhen:       "kong",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, tc.runWhen, ">=3.0.0")
			setup(t)
			ctx := context.Background()
			require.NoError(t, apply(ctx, tc.firstFile))
			require.NoError(t, apply(ctx, tc.secondFile))

			out, _ := dump()

			expected, err := readFile(tc.expectedState)
			if err != nil {
				t.Fatalf("failed to read expected state: %v", err)
			}

			assert.Equal(t, expected, out)
		})
	}

	t.Run("updates existing entities", func(t *testing.T) {
		runWhen(t, "kong", ">=3.0.0")
		setup(t)

		apply("testdata/apply/007-update-existing-entity/service-01.yaml")

		out, err := dump()
		require.NoError(t, err)
		expectedOriginal, err := readFile("testdata/apply/007-update-existing-entity/expected-state-01.yaml")
		require.NoError(t, err, "failed to read expected state")

		assert.Equal(t, expectedOriginal, out)

		apply("testdata/apply/007-update-existing-entity/service-02.yaml")
		expectedChanged, err := readFile("testdata/apply/007-update-existing-entity/expected-state-02.yaml")
		require.NoError(t, err, "failed to read expected state")

		outChanged, err := dump()
		require.NoError(t, err)
		assert.Equal(t, expectedChanged, outChanged)
	})
}
