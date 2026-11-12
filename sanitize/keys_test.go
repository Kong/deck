package sanitize

import (
	"context"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/require"
)

func Test_SanitizeKeyOptionalPEM(t *testing.T) {
	privateKeyPEM, privateKey, err := generateRSAPrivateKeyPEM()
	require.NoError(t, err)
	publicKey := generateRSAPublicKeyPEM(privateKey)

	for name, pem := range map[string]*kong.PEM{
		"public only":  {PublicKey: &publicKey},
		"private only": {PrivateKey: &privateKeyPEM},
		"both":         {PublicKey: &publicKey, PrivateKey: &privateKeyPEM},
	} {
		t.Run(name, func(t *testing.T) {
			sanitizer := NewSanitizer(&SanitizerOptions{
				Ctx: context.Background(),
				Content: &file.Content{Keys: []file.FKey{{Key: kong.Key{
					KID: kong.String("test-key"), PEM: pem,
				}}}},
			})
			var result *file.Content
			require.NotPanics(t, func() { result, err = sanitizer.Sanitize() })
			require.NoError(t, err)
			require.Len(t, result.Keys, 1)
			require.NotNil(t, result.Keys[0].PEM)
			if pem.PublicKey != nil {
				require.NotEqual(t, *pem.PublicKey, *result.Keys[0].PEM.PublicKey)
				require.Equal(t, *result.Keys[0].PEM.PublicKey, sanitizer.sanitizedMap[*pem.PublicKey])
			} else {
				require.Nil(t, result.Keys[0].PEM.PublicKey)
			}
			if pem.PrivateKey != nil {
				require.NotEqual(t, *pem.PrivateKey, *result.Keys[0].PEM.PrivateKey)
				require.Equal(t, *result.Keys[0].PEM.PrivateKey, sanitizer.sanitizedMap[*pem.PrivateKey])
			} else {
				require.Nil(t, result.Keys[0].PEM.PrivateKey)
			}
		})
	}
}
