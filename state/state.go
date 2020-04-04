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
	SNIs           *SNIsCollection
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
	basicAuthTemp := newBasicAuthsCollection(collection{})
	jwtAuthTemp := newJWTAuthsCollection(collection{})
	oauth2CredsTemp := newOauth2CredsCollection(collection{})

	var schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			serviceTableName:     serviceTableSchema,
			routeTableName:       routeTableSchema,
			upstreamTableName:    upstreamTableSchema,
			targetTableName:      targetTableSchema,
			certificateTableName: certificateTableSchema,
			sniTableName:         sniTableSchema,
			caCertTableName:      caCertTableSchema,
			pluginTableName:      pluginTableSchema,
			consumerTableName:    consumerTableSchema,

			keyAuthTemp.TableName():     keyAuthTemp.Schema(),
			hmacAuthTemp.TableName():    hmacAuthTemp.Schema(),
			basicAuthTemp.TableName():   basicAuthTemp.Schema(),
			jwtAuthTemp.TableName():     jwtAuthTemp.Schema(),
			oauth2CredsTemp.TableName(): oauth2CredsTemp.Schema(),

			aclGroupTableName: aclGroupTableSchema,
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
	state.SNIs = (*SNIsCollection)(&state.common)
	state.CACertificates = (*CACertificatesCollection)(&state.common)
	state.Plugins = (*PluginsCollection)(&state.common)
	state.Consumers = (*ConsumersCollection)(&state.common)

	state.KeyAuths = newKeyAuthsCollection(state.common)
	state.HMACAuths = newHMACAuthsCollection(state.common)
	state.BasicAuths = newBasicAuthsCollection(state.common)
	state.JWTAuths = newJWTAuthsCollection(state.common)
	state.Oauth2Creds = newOauth2CredsCollection(state.common)

	state.ACLGroups = (*ACLGroupsCollection)(&state.common)
	return &state, nil
}
