package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/pkg/errors"
)

type collection struct {
	db *memdb.MemDB
}

// KongState is an in-memory database representation
// of Kong's configuration.
type KongState struct {
	common         collection
	Services       *ServicesCollection
	Routes         *RoutesCollection
	Upstreams      *UpstreamsCollection
	Targets        *TargetsCollection
	Certificates   *CertificatesCollection
	CACertificates *CACertificatesCollection
	Plugins        *PluginsCollection
	Consumers      *ConsumersCollection

	KeyAuths    *KeyAuthsCollection
	HMACAuths   *HMACAuthsCollection
	JWTAuths    *JWTAuthsCollection
	BasicAuths  *BasicAuthsCollection
	ACLGroups   *ACLGroupsCollection
	Oauth2Creds *Oauth2CredsCollection
}

// NewKongState creates a new in-memory KongState.
func NewKongState() (*KongState, error) {

	// TODO FIXME clean up the mess
	keyAuthTemp := newKeyAuthsCollection(collection{})
	hmacAuthTemp := newHMACAuthsCollection(collection{})

	var schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			serviceTableName:     serviceTableSchema,
			routeTableName:       routeTableSchema,
			upstreamTableName:    upstreamTableSchema,
			targetTableName:      targetTableSchema,
			certificateTableName: certificateTableSchema,
			caCertTableName:      caCertTableSchema,
			pluginTableName:      pluginTableSchema,
			consumerTableName:    consumerTableSchema,

			keyAuthTemp.TableName():  keyAuthTemp.Schema(),
			hmacAuthTemp.TableName(): hmacAuthTemp.Schema(),
			basicAuthTableName:       basicAuthTableSchema,
			jwtAuthTableName:         jwtAuthTableSchema,
			oauth2CredTableName:      oauth2CredTableSchema,
			aclGroupTableName:        aclGroupTableSchema,
		},
	}

	memDB, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, errors.Wrap(err, "creating new ServiceCollection")
	}
	var state KongState
	state.common = collection{
		db: memDB,
	}

	state.Services = (*ServicesCollection)(&state.common)
	state.Routes = (*RoutesCollection)(&state.common)
	state.Upstreams = (*UpstreamsCollection)(&state.common)
	state.Targets = (*TargetsCollection)(&state.common)
	state.Certificates = (*CertificatesCollection)(&state.common)
	state.CACertificates = (*CACertificatesCollection)(&state.common)
	state.Plugins = (*PluginsCollection)(&state.common)
	state.Consumers = (*ConsumersCollection)(&state.common)

	state.KeyAuths = newKeyAuthsCollection(state.common)
	state.HMACAuths = newHMACAuthsCollection(state.common)
	state.JWTAuths = (*JWTAuthsCollection)(&state.common)
	state.BasicAuths = (*BasicAuthsCollection)(&state.common)
	state.ACLGroups = (*ACLGroupsCollection)(&state.common)
	state.Oauth2Creds = (*Oauth2CredsCollection)(&state.common)
	return &state, nil
}
