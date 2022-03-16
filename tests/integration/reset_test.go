//go:build integration

package integration

import (
	"testing"

	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

var caCert = &kong.CACertificate{
	CertDigest: kong.String("b865971cecadd7bac9487901c9269c1fa903b3a3b521a927c5e2513f692ac61e"),
	Cert: kong.String(`-----BEGIN CERTIFICATE-----
MIICvDCCAaSgAwIBAgIJAID17vZt1yWyMA0GCSqGSIb3DQEBCwUAMBMxETAPBgNV
BAMMCEhlbGxvTmV3MB4XDTIyMDMxNTE5MTgzOVoXDTIyMDQxNDE5MTgzOVowEzER
MA8GA1UEAwwISGVsbG9OZXcwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIB
AQDL1l6EB2rDPjKDoJX52VsnO8bdKnoGq2s5Er7piVQjYBA6U7NzEJwvsaL9uG1p
/OFud8uwJFCm0NF1DxkNA+qpUvaBXBnn4htbXE20C7HwAWCUU0TUWgTpGYC0EkGZ
VlbXoQ1SewK+AERjdBKqa0U9Wk0gkD0kVc2UfO7rxU7w6nkoFPgBI1IlJZXM5TVg
1AeJDrdgUSa/fsja5qOYVcwGiUgqMr3nMs1jBJRgwhC0ELF1lFaANouqPC4KweLE
FNgam69AZallFNZOKVh6vJLKBfE9I8TM5yBpRllhKAaUv1qWlPFYxoIPvnFzQPku
ExGbYR6asSXwq6UHxREIOno1AgMBAAGjEzARMA8GA1UdEwQIMAYBAf8CAQAwDQYJ
KoZIhvcNAQELBQADggEBAHXc6lc6BGzdjwWX8XViBxY1NnK5HxNfD+rP7/JJ4m33
zoTteY+KRcKo6t49TDqfpnVfCGunnoGOFP5ATY29vUavigICw7SGGLKWIM38c0bH
bx14/d/LQd2LaNd/cemTDkF3XJi3OdrGJPNOVLfX0InqbmwBzariWCwzufwHGxwR
WpOh8Qv2kFPuFVwlQNPNMhV7qsa/NM77Wo4Q6kA5V3aYSnF+KbWF3by/SqUC5JMz
cbvPj0Yzt97v7FpOILcDcMWjxuUnvuUYvGuB5tzBEe91s3ZTUK0A5moYOYkTHUlX
9CkGSwFE+jBTxUBPKzm3MVoK2cGoX8gEpzcYSwjM8Ws=
-----END CERTIFICATE-----`),
}

func Test_Reset_SkipCACert(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	if err != nil {
		t.Errorf(err.Error())
	}

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "reset with --skip-ca-certificates should ignore CA certs",
			kongFile: "testdata/reset/001-skip-ca-cert/kong.yaml",
			expectedState: utils.KongRawState{
				CACertificates: []*kong.CACertificate{caCert},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// ca_certificates first appeared in 1.3, but we limit to 2.7+
			// here because the schema changed and the entities aren't the same
			// across all versions, even though the skip functionality works the same.
			kong.RunWhenKong(t, ">=2.7.0")
			teardown := setup(t)
			defer teardown(t)

			sync(tc.kongFile)
			reset(t, "--skip-ca-certificates")
			testKongState(t, client, tc.expectedState, nil)
		})
	}
}
