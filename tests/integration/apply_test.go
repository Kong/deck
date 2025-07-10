//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
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
		{
			name:          "accepts route foreign keys",
			firstFile:     "testdata/apply/006-foreign-keys-routes/route-01.yaml",
			secondFile:    "testdata/apply/006-foreign-keys-routes/plugin-01.yaml",
			expectedState: "testdata/apply/006-foreign-keys-routes/expected-state.yaml",
			runWhen:       "kong",
		},
		{
			name:          "accepts route updates",
			firstFile:     "testdata/apply/008-update-existing-nested-entity/route-01.yaml",
			secondFile:    "testdata/apply/008-update-existing-nested-entity/route-02.yaml",
			expectedState: "testdata/apply/008-update-existing-nested-entity/route-expected-state.yaml",
			runWhen:       "kong",
		},
		{
			name:          "accepts consumer group consumer updates",
			firstFile:     "testdata/apply/008-update-existing-nested-entity/consumer-group-01.yaml",
			secondFile:    "testdata/apply/008-update-existing-nested-entity/consumer-group-02.yaml",
			expectedState: "testdata/apply/008-update-existing-nested-entity/consumer-group-expected-state.yaml",
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

		err := apply(context.Background(), "testdata/apply/007-update-existing-entity/service-01.yaml")
		require.NoError(t, err, "failed to apply service-01")

		out, err := dump()
		require.NoError(t, err)
		expectedOriginal, err := readFile("testdata/apply/007-update-existing-entity/expected-state-01.yaml")
		require.NoError(t, err, "failed to read expected state")

		assert.Equal(t, expectedOriginal, out)

		err = apply(context.Background(), "testdata/apply/007-update-existing-entity/service-02.yaml")
		require.NoError(t, err, "failed to apply service-02")

		expectedChanged, err := readFile("testdata/apply/007-update-existing-entity/expected-state-02.yaml")
		require.NoError(t, err, "failed to read expected state")

		outChanged, err := dump()
		require.NoError(t, err)
		assert.Equal(t, expectedChanged, outChanged)
	})
}

// test scope:
//   - enterprise: >=3.0.0
//   - konnect
func Test_Apply_Custom_Entities(t *testing.T) {
	runWhenEnterpriseOrKonnect(t, ">=3.0.0")
	setup(t)

	ctx := context.Background()
	tests := []struct {
		name                   string
		initialStateFile       string
		targetPartialStateFile string
	}{
		{
			name:                   "degraphql routes",
			initialStateFile:       "testdata/apply/008-custom-entities/initial-state.yaml",
			targetPartialStateFile: "testdata/apply/008-custom-entities/partial-update.yaml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				reset(t)
			})
			err := sync(ctx, tc.initialStateFile)
			require.NoError(t, err)

			err = apply(ctx, tc.targetPartialStateFile)
			require.NoError(t, err)
		})
	}
}

// test scope:
//   - >=3.1.0
func Test_Apply_KeysAndKeySets(t *testing.T) {
	runWhen(t, "kong", ">=3.1.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)
	ctx := context.Background()

	tests := []struct {
		name             string
		initialStateFile string
		updateStateFile  string
		expectedState    utils.KongRawState
	}{
		{
			name:             "keys and key_sets",
			initialStateFile: "testdata/apply/009-keys-and-key_sets/initial.yaml",
			updateStateFile:  "testdata/apply/009-keys-and-key_sets/update.yaml",
			expectedState: utils.KongRawState{
				Keys: []*kong.Key{
					{
						ID:   kong.String("f21a7073-1183-4b1c-bd87-4d5b8b18eeb4"),
						Name: kong.String("foo"),
						KID:  kong.String("vsR8NCNV_1_LB06LqudGa2r-T0y4Z6VQVYue9IQz6A4"),
						Set: &kong.KeySet{
							ID: kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935453"),
						},
						JWK: kong.String("{\"kty\": \"RSA\", \"kid\": \"vsR8NCNV_1_LB06LqudGa2r-T0y4Z6VQVYue9IQz6A4\", \"n\": \"v2KAzzfruqctVHaE9WSCWIg1xAhMwxTIK-i56WNqPtpWBo9AqxcVea8NyVctEjUNq_mix5CklNy3ru7ARh7rBG_LU65fzs4fY_uYalul3QZSnr61Gj-cTUB3Gy4PhA63yXCbYRR3gDy6WR_wfis1MS61j0R_AjgXuVufmmC0F7R9qSWfR8ft0CbQgemEHY3ddKeW7T7fKv1jnRwYAkl5B_xtvxRFIYT-uR9NNftixNpUIW7q8qvOH7D9icXOg4_wIVxTRe5QiRYwEFoUbV1V9bFtu5FLal0vZnLaWwg5tA6enhzBpxJNdrS0v1RcPpyeNP-9r3cUDGmeftwz9v95UQ\", \"e\": \"AQAB\", \"alg\": \"A256GCM\"}"), //nolint:lll
					},
					{
						ID:   kong.String("d7cef208-23c3-46f8-94e8-fa1eddf43f0a"),
						Name: kong.String("baz"),
						KID:  kong.String("IiI4ffge7LZXPztrZVOt26zgRt0EPsWPaxAmwhbJhDQ"),
						Set: &kong.KeySet{
							ID: kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935345"),
						},
						JWK: kong.String("{\n      \"kty\": \"RSA\",\n      \"kid\": \"IiI4ffge7LZXPztrZVOt26zgRt0EPsWPaxAmwhbJhDQ\",\n      \"use\": \"sig\",\n      \"alg\": \"RS256\",\n      \"e\": \"AQAB\",\n      \"n\": \"1Sn1X_y-RUzGna0hR00Wu64ZtY5N5BVzpRIby9wQ5EZVyWL9DRhU5PXqM3Y5gzgUVEQu548qQcMKOfs46PhOQudz-HPbwKWzcJCDUeNQsxdAEhW1uJR0EEV_SGJ-jTuKGqoEQc7bNrmhyXBMIeMkTeE_-ys75iiwvNjYphiOhsokC_vRTf_7TOPTe1UQasgxEVSLlTsen0vtK_FXcpbwdxZt02IysICcX5TcWX_XBuFP4cpwI9AS3M-imc01awc1t7FE5UWp62H5Ro2S5V9YwdxSjf4lX87AxYmawaWAjyO595XLuIXA3qt8-irzbCeglR1-cTB7a4I7_AclDmYrpw\"\n  }"), //nolint:lll
					},
					{
						ID:   kong.String("03ad4618-82bb-4375-b9d1-edeefced868d"),
						Name: kong.String("my-pem-key"),
						KID:  kong.String("my-pem-key"),
						Set: &kong.KeySet{
							ID: kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935345"),
						},
						PEM: &kong.PEM{
							PublicKey:  kong.String("-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqvxMU4LTcHBYmCuLMhMP\nDWlZdcNRXuJkw26MRjLBxXjnPAyDolmuFFMIqPDlSaJkkzu2tn7m9p8KB90wLiMC\nIbDjseruCO+7EaIRY4d6RdpE+XowCjJu7SbC2CqWBAzKkO7WWAunO3KOsQRk1NEK\nI51CoZ26LPYQvjIGIY2/pPxq0Ydl9dyURqVfmTywni1WeScgdEZXuy9WIcobqBST\n8vV5Q5HJsZNFLR7Fy61+HHfnQiWIYyi6h8QRT+Css9y5KbH7KuN6tnb94UZaOmHl\nYeoHcP/CqviZnQOf5804qcVpPKbsGU8jupTriiJZU3a8f59eHV0ybI4ORXYgDSWd\nFQIDAQAB\n-----END PUBLIC KEY-----"),                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               //nolint:lll
							PrivateKey: kong.String("-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEAqvxMU4LTcHBYmCuLMhMPDWlZdcNRXuJkw26MRjLBxXjnPAyD\nolmuFFMIqPDlSaJkkzu2tn7m9p8KB90wLiMCIbDjseruCO+7EaIRY4d6RdpE+Xow\nCjJu7SbC2CqWBAzKkO7WWAunO3KOsQRk1NEKI51CoZ26LPYQvjIGIY2/pPxq0Ydl\n9dyURqVfmTywni1WeScgdEZXuy9WIcobqBST8vV5Q5HJsZNFLR7Fy61+HHfnQiWI\nYyi6h8QRT+Css9y5KbH7KuN6tnb94UZaOmHlYeoHcP/CqviZnQOf5804qcVpPKbs\nGU8jupTriiJZU3a8f59eHV0ybI4ORXYgDSWdFQIDAQABAoIBAEOOqAGfATe9y+Nj\n4P2J9jqQU15qK65XuQRWm2npCBKj8IkTULdGw7cYD6XgeFedqCtcPpbgkRUERYxR\n4oV4I5F4OJ7FegNh5QHUjRZMIw2Sbgo8Mtr0jkt5MycBvIAhJbAaDep/wDWGz8Y1\nPDmx1lW3/umoTjURjA/5594+CWiABYzuIi4WprWe4pIKqSKOMHnCYVAD243mwJ7y\nvsatO3LRKYfLw74ifCYhWNBHaZwfw+OO2P5Ku0AGhY4StOLCHobJ8/KkkmkTlYzv\nrcF4cVdvpBfdTEQed0oD7u3xfnp3GpNU3wZFsZJRSVXouhroaMC7en4uMc+5yguW\nqrPIoEkCgYEAxm1UllY9rRfGV6884hdBFKDjE825BC1VlqcRIUEB4CpJvUF/6+A3\ngx5c4nKDJAFQMrWpr4jOcq3iLiWnJ73e80b+JpWFODdt16g2KCOINs1j8vf2U6Og\nx+Vo8vHek/Uomz1n5W0oXrJ4VedHl9NYa8r/YrVXd4k4WcaA0TXmMhMCgYEA3Jit\nzrEmrQIrLK66RgXF2RafA5c3atRHWBb5ddnGk0bV90cfsTsaDMDvpy7ZYgojBNpw\n7U6AYzqnPro6cHEginV97BFb6oetMvOWvljUob+tpnYOofgwk2hw7PeChViX7iS9\nujgTygi8ZIc2G0r7xntH+v6WHKp4yNQiCAyfGTcCgYAYKgZMDJKUOrn3wapraiON\nzI36wmnOnWq33v6SCyWcU+oI9yoJ4pNAD3mGRiW8Q8CtfDv+2W0ywAQ0VHeHunKl\nM7cNodXIY8+nnJ+Dwdf7vIV4eEPyKZIR5dkjBNtzLz7TsOWvJdzts1Q+Od0ZGy7A\naccyER1mvDo1jJvxXlv7KwKBgQDDBK9TdUVt2eb1X5sJ4HyiiN8XO44ggX55IAZ1\n64skFJGARH5+HnPPJpo3wLEpfTCsT7lZ8faKwwWr7NNRKJHOFkS2eDo8QqoZy0NP\nEBUa0evgp6oUAuheyQxcUgwver0GKbEZeg30pHh4nxh0VHv1YnOmL3/h48tYMEHN\nv+q/TQKBgQCXQmN8cY2K7UfZJ6BYEdguQZS5XISFbLNkG8wXQX9vFiF8TuSWawDN\nTrRHVDGwoMGWxjZBLCsitA6zwrMLJZs4RuetKHFou7MiDQ69YGdfNRlRvD5QCJDc\nY0ICsYjI7VM89Qj/41WQyRHYHm7E9key3avMGdbYtxdc0Ku4LnD4zg==\n-----END RSA PRIVATE KEY-----"), //nolint:lll
						},
					},
				},
				KeySets: []*kong.KeySet{
					{
						Name: kong.String("bar"),
						ID:   kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935453"),
					},
					{
						Name: kong.String("bar-new"),
						ID:   kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935345"),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			err := sync(ctx, tc.initialStateFile)
			require.NoError(t, err)

			err = apply(ctx, tc.updateStateFile)
			require.NoError(t, err)

			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - konnect
func Test_Apply_KeysAndKeySets_Konnect(t *testing.T) {
	setDefaultKonnectControlPlane(t)
	runWhenKonnect(t)
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)
	ctx := context.Background()

	tests := []struct {
		name             string
		initialStateFile string
		updateStateFile  string
		expectedState    utils.KongRawState
	}{
		{
			name:             "keys and key_sets",
			initialStateFile: "testdata/apply/009-keys-and-key_sets/initial.yaml",
			updateStateFile:  "testdata/apply/009-keys-and-key_sets/update.yaml",
			expectedState: utils.KongRawState{
				Keys: []*kong.Key{
					{
						ID:   kong.String("f21a7073-1183-4b1c-bd87-4d5b8b18eeb4"),
						Name: kong.String("foo"),
						KID:  kong.String("vsR8NCNV_1_LB06LqudGa2r-T0y4Z6VQVYue9IQz6A4"),
						Set: &kong.KeySet{
							ID: kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935453"),
						},
						JWK: kong.String("{\"kid\":\"vsR8NCNV_1_LB06LqudGa2r-T0y4Z6VQVYue9IQz6A4\",\"kty\":\"RSA\",\"alg\":\"A256GCM\",\"n\":\"v2KAzzfruqctVHaE9WSCWIg1xAhMwxTIK-i56WNqPtpWBo9AqxcVea8NyVctEjUNq_mix5CklNy3ru7ARh7rBG_LU65fzs4fY_uYalul3QZSnr61Gj-cTUB3Gy4PhA63yXCbYRR3gDy6WR_wfis1MS61j0R_AjgXuVufmmC0F7R9qSWfR8ft0CbQgemEHY3ddKeW7T7fKv1jnRwYAkl5B_xtvxRFIYT-uR9NNftixNpUIW7q8qvOH7D9icXOg4_wIVxTRe5QiRYwEFoUbV1V9bFtu5FLal0vZnLaWwg5tA6enhzBpxJNdrS0v1RcPpyeNP-9r3cUDGmeftwz9v95UQ\",\"e\":\"AQAB\"}"), //nolint:lll
					},
					{
						ID:   kong.String("d7cef208-23c3-46f8-94e8-fa1eddf43f0a"),
						Name: kong.String("baz"),
						KID:  kong.String("IiI4ffge7LZXPztrZVOt26zgRt0EPsWPaxAmwhbJhDQ"),
						Set: &kong.KeySet{
							ID: kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935345"),
						},
						JWK: kong.String("{\"kid\":\"IiI4ffge7LZXPztrZVOt26zgRt0EPsWPaxAmwhbJhDQ\",\"kty\":\"RSA\",\"use\":\"sig\",\"alg\":\"RS256\",\"n\":\"1Sn1X_y-RUzGna0hR00Wu64ZtY5N5BVzpRIby9wQ5EZVyWL9DRhU5PXqM3Y5gzgUVEQu548qQcMKOfs46PhOQudz-HPbwKWzcJCDUeNQsxdAEhW1uJR0EEV_SGJ-jTuKGqoEQc7bNrmhyXBMIeMkTeE_-ys75iiwvNjYphiOhsokC_vRTf_7TOPTe1UQasgxEVSLlTsen0vtK_FXcpbwdxZt02IysICcX5TcWX_XBuFP4cpwI9AS3M-imc01awc1t7FE5UWp62H5Ro2S5V9YwdxSjf4lX87AxYmawaWAjyO595XLuIXA3qt8-irzbCeglR1-cTB7a4I7_AclDmYrpw\",\"e\":\"AQAB\"}"), //nolint:lll
					},
					{
						ID:   kong.String("03ad4618-82bb-4375-b9d1-edeefced868d"),
						Name: kong.String("my-pem-key"),
						KID:  kong.String("my-pem-key"),
						Set: &kong.KeySet{
							ID: kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935345"),
						},
						PEM: &kong.PEM{
							PublicKey:  kong.String("-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqvxMU4LTcHBYmCuLMhMP\nDWlZdcNRXuJkw26MRjLBxXjnPAyDolmuFFMIqPDlSaJkkzu2tn7m9p8KB90wLiMC\nIbDjseruCO+7EaIRY4d6RdpE+XowCjJu7SbC2CqWBAzKkO7WWAunO3KOsQRk1NEK\nI51CoZ26LPYQvjIGIY2/pPxq0Ydl9dyURqVfmTywni1WeScgdEZXuy9WIcobqBST\n8vV5Q5HJsZNFLR7Fy61+HHfnQiWIYyi6h8QRT+Css9y5KbH7KuN6tnb94UZaOmHl\nYeoHcP/CqviZnQOf5804qcVpPKbsGU8jupTriiJZU3a8f59eHV0ybI4ORXYgDSWd\nFQIDAQAB\n-----END PUBLIC KEY-----"),                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               //nolint:lll
							PrivateKey: kong.String("-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEAqvxMU4LTcHBYmCuLMhMPDWlZdcNRXuJkw26MRjLBxXjnPAyD\nolmuFFMIqPDlSaJkkzu2tn7m9p8KB90wLiMCIbDjseruCO+7EaIRY4d6RdpE+Xow\nCjJu7SbC2CqWBAzKkO7WWAunO3KOsQRk1NEKI51CoZ26LPYQvjIGIY2/pPxq0Ydl\n9dyURqVfmTywni1WeScgdEZXuy9WIcobqBST8vV5Q5HJsZNFLR7Fy61+HHfnQiWI\nYyi6h8QRT+Css9y5KbH7KuN6tnb94UZaOmHlYeoHcP/CqviZnQOf5804qcVpPKbs\nGU8jupTriiJZU3a8f59eHV0ybI4ORXYgDSWdFQIDAQABAoIBAEOOqAGfATe9y+Nj\n4P2J9jqQU15qK65XuQRWm2npCBKj8IkTULdGw7cYD6XgeFedqCtcPpbgkRUERYxR\n4oV4I5F4OJ7FegNh5QHUjRZMIw2Sbgo8Mtr0jkt5MycBvIAhJbAaDep/wDWGz8Y1\nPDmx1lW3/umoTjURjA/5594+CWiABYzuIi4WprWe4pIKqSKOMHnCYVAD243mwJ7y\nvsatO3LRKYfLw74ifCYhWNBHaZwfw+OO2P5Ku0AGhY4StOLCHobJ8/KkkmkTlYzv\nrcF4cVdvpBfdTEQed0oD7u3xfnp3GpNU3wZFsZJRSVXouhroaMC7en4uMc+5yguW\nqrPIoEkCgYEAxm1UllY9rRfGV6884hdBFKDjE825BC1VlqcRIUEB4CpJvUF/6+A3\ngx5c4nKDJAFQMrWpr4jOcq3iLiWnJ73e80b+JpWFODdt16g2KCOINs1j8vf2U6Og\nx+Vo8vHek/Uomz1n5W0oXrJ4VedHl9NYa8r/YrVXd4k4WcaA0TXmMhMCgYEA3Jit\nzrEmrQIrLK66RgXF2RafA5c3atRHWBb5ddnGk0bV90cfsTsaDMDvpy7ZYgojBNpw\n7U6AYzqnPro6cHEginV97BFb6oetMvOWvljUob+tpnYOofgwk2hw7PeChViX7iS9\nujgTygi8ZIc2G0r7xntH+v6WHKp4yNQiCAyfGTcCgYAYKgZMDJKUOrn3wapraiON\nzI36wmnOnWq33v6SCyWcU+oI9yoJ4pNAD3mGRiW8Q8CtfDv+2W0ywAQ0VHeHunKl\nM7cNodXIY8+nnJ+Dwdf7vIV4eEPyKZIR5dkjBNtzLz7TsOWvJdzts1Q+Od0ZGy7A\naccyER1mvDo1jJvxXlv7KwKBgQDDBK9TdUVt2eb1X5sJ4HyiiN8XO44ggX55IAZ1\n64skFJGARH5+HnPPJpo3wLEpfTCsT7lZ8faKwwWr7NNRKJHOFkS2eDo8QqoZy0NP\nEBUa0evgp6oUAuheyQxcUgwver0GKbEZeg30pHh4nxh0VHv1YnOmL3/h48tYMEHN\nv+q/TQKBgQCXQmN8cY2K7UfZJ6BYEdguQZS5XISFbLNkG8wXQX9vFiF8TuSWawDN\nTrRHVDGwoMGWxjZBLCsitA6zwrMLJZs4RuetKHFou7MiDQ69YGdfNRlRvD5QCJDc\nY0ICsYjI7VM89Qj/41WQyRHYHm7E9key3avMGdbYtxdc0Ku4LnD4zg==\n-----END RSA PRIVATE KEY-----"), //nolint:lll
						},
					},
				},
				KeySets: []*kong.KeySet{
					{
						Name: kong.String("bar"),
						ID:   kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935453"),
					},
					{
						Name: kong.String("bar-new"),
						ID:   kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935345"),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			err := sync(ctx, tc.initialStateFile)
			require.NoError(t, err)

			err = apply(ctx, tc.updateStateFile)
			require.NoError(t, err)

			testKongState(t, client, true, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - konnect
func Test_Apply_NestedEntities_Konnect(t *testing.T) {
	setDefaultKonnectControlPlane(t)
	runWhenKonnect(t)
	setup(t)

	ctx := context.Background()

	tests := []struct {
		name          string
		firstFile     string
		secondFile    string
		expectedState string
	}{
		{
			name:          "accepts route updates",
			firstFile:     "testdata/apply/008-update-existing-nested-entity/route-01.yaml",
			secondFile:    "testdata/apply/008-update-existing-nested-entity/route-02.yaml",
			expectedState: "testdata/apply/008-update-existing-nested-entity/route-expected-state.yaml",
		},
		{
			name:          "accepts consumer group consumer updates",
			firstFile:     "testdata/apply/008-update-existing-nested-entity/consumer-group-01.yaml",
			secondFile:    "testdata/apply/008-update-existing-nested-entity/consumer-group-02.yaml",
			expectedState: "testdata/apply/008-update-existing-nested-entity/consumer-group-expected-state.yaml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			err := sync(ctx, tc.firstFile)
			require.NoError(t, err)

			err = apply(ctx, tc.secondFile)
			require.NoError(t, err)

			out, _ := dump()

			expected, err := readFile(tc.expectedState)
			if err != nil {
				t.Fatalf("failed to read expected state: %v", err)
			}

			assert.Equal(t, expected, out)
		})
	}
}
