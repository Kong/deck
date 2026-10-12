package cmd

import (
	"testing"

	"github.com/kong/deck/kong2tf"
	"github.com/stretchr/testify/require"
)

func TestKong2TfProviderFlag(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		generate    string
		wantErr     string
		wantDefault string
	}{
		{
			name:        "defaults to konnect",
			wantDefault: string(kong2tf.ProviderKonnect),
		},
		{
			name:     "accepts konnect",
			provider: string(kong2tf.ProviderKonnect),
		},
		{
			name:     "accepts kong gateway",
			provider: string(kong2tf.ProviderKongGateway),
		},
		{
			name:     "rejects unknown provider",
			provider: "unknown",
			wantErr:  "invalid value 'unknown' found for the 'provider' flag",
		},
		{
			name:     "rejects control plane imports for kong gateway",
			provider: string(kong2tf.ProviderKongGateway),
			generate: "cp-id",
			wantErr:  "--generate-imports-for-control-plane-id can only be used with --provider konnect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newKong2TfCmd()
			if tt.provider != "" {
				require.NoError(t, cmd.Flags().Set("provider", tt.provider))
			}
			if tt.generate != "" {
				require.NoError(t, cmd.Flags().Set("generate-imports-for-control-plane-id", tt.generate))
			}

			provider, err := cmd.Flags().GetString("provider")
			require.NoError(t, err)
			if tt.wantDefault != "" {
				require.Equal(t, tt.wantDefault, provider)
			}

			err = cmd.PreRunE(cmd, nil)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
