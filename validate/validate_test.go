package validate

import (
	"testing"

	"github.com/kong/go-database-reconciler/pkg/konnect"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func Test_getEntityNameOrID(t *testing.T) {
	tests := []struct {
		name     string
		entity   interface{}
		expected string
	}{
		{
			name: "get service name",
			entity: &kong.Service{
				Name: kong.String("svc1"),
			},
			expected: "svc1",
		},
		{
			name: "get route name",
			entity: &kong.Route{
				Name: kong.String("route1"),
			},
			expected: "route1",
		},
		{
			name: "get consumer ID",
			entity: &kong.Consumer{
				ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Username: kong.String("foo"),
			},
			expected: "4bfcb11f-c962-4817-83e5-9433cf20b663",
		},
		{
			name: "get key-auth ID",
			entity: &kong.KeyAuth{
				ID:  kong.String("5f1ef1ea-a2a5-4a1b-adbb-b0d3434013e5"),
				Key: kong.String("foo-apikey"),
				Consumer: &kong.Consumer{
					ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				},
			},
			expected: "5f1ef1ea-a2a5-4a1b-adbb-b0d3434013e5",
		},
		{
			name: "get basic-auth ID",
			entity: &kong.BasicAuth{
				ID:       kong.String("92f4c849-960b-43af-aad3-f307051408d3"),
				Username: kong.String("basic-username"),
				Password: kong.String("basic-password"),
				Consumer: &kong.Consumer{
					ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				},
			},
			expected: "92f4c849-960b-43af-aad3-f307051408d3",
		},
		{
			name: "get jwt-auth ID",
			entity: &kong.JWTAuth{
				ID:     kong.String("917b9402-1be0-49d2-b482-ca4dccc2054e"),
				Key:    kong.String("jwt-key"),
				Secret: kong.String("jwt-secret"),
				Consumer: &kong.Consumer{
					ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				},
			},
			expected: "917b9402-1be0-49d2-b482-ca4dccc2054e",
		},
		{
			name: "get hmac-auth ID",
			entity: &kong.HMACAuth{
				ID:       kong.String("e5d81b73-bf9e-42b0-9d68-30a1d791b9c9"),
				Username: kong.String("hmac-username"),
				Secret:   kong.String("hmac-secret"),
				Consumer: &kong.Consumer{
					ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				},
			},
			expected: "e5d81b73-bf9e-42b0-9d68-30a1d791b9c9",
		},
		{
			name: "get acl-group ID",
			entity: &kong.ACLGroup{
				ID:    kong.String("b7c9352a-775a-4ba5-9869-98e926a3e6cb"),
				Group: kong.String("foo-group"),
				Consumer: &kong.Consumer{
					ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				},
			},
			expected: "b7c9352a-775a-4ba5-9869-98e926a3e6cb",
		},
		{
			name: "get oauth2 name",
			entity: &kong.Oauth2Credential{
				ID:       kong.String("4eef5285-3d6a-4f6b-b659-8957a940e2ca"),
				ClientID: kong.String("oauth2-clientid"),
				Name:     kong.String("oauth2-name"),
				Consumer: &kong.Consumer{
					ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				},
			},
			expected: "oauth2-name",
		},
		{
			name: "get mtls-auth ID",
			entity: &kong.MTLSAuth{
				ID:          kong.String("92f4c829-968b-42af-afd3-f337051508d3"),
				SubjectName: kong.String("test@example.com"),
				Consumer: &kong.Consumer{
					ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				},
			},
			expected: "92f4c829-968b-42af-afd3-f337051508d3",
		},
		{
			name: "get upstream name",
			entity: &kong.Upstream{
				ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Name: kong.String("foo"),
			},
			expected: "foo",
		},
		{
			name: "get cert ID",
			entity: &kong.Certificate{
				ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Cert: kong.String("foo"),
				Key:  kong.String("bar"),
			},
			expected: "4bfcb11f-c962-4817-83e5-9433cf20b663",
		},
		{
			name: "get CA cert ID",
			entity: &kong.CACertificate{
				ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Cert: kong.String("foo"),
			},
			expected: "4bfcb11f-c962-4817-83e5-9433cf20b663",
		},
		{
			name: "get plugin name",
			entity: &kong.Plugin{
				ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Name: kong.String("foo"),
			},
			expected: "foo",
		},
		{
			name: "get target ID",
			entity: &kong.Target{
				ID:     kong.String("f7e64af5-e438-4a9b-8ff8-ec6f5f06dccb"),
				Target: kong.String("bar"),
				Upstream: &kong.Upstream{
					ID: kong.String("f77ca8c7-581d-45a4-a42c-c003234228e1"),
				},
			},
			expected: "f7e64af5-e438-4a9b-8ff8-ec6f5f06dccb",
		},
		{
			name: "get service package name",
			entity: &konnect.ServicePackage{
				ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Name: kong.String("foo"),
			},
			expected: "foo",
		},
		{
			name: "get document ID",
			entity: &konnect.Document{
				ID:        kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Path:      kong.String("/foo.md"),
				Published: kong.Bool(true),
				Content:   kong.String("foo"),
			},
			expected: "4bfcb11f-c962-4817-83e5-9433cf20b663",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEntityNameOrID(tt.entity)
			assert.Equal(t, tt.expected, got)
		})
	}
}
