package sanitize

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/kong/go-kong/kong"
)

const (
	Plugin  = "Plugin"
	Partial = "Partial"
)

var entityMap = map[string]string{
	"ACLGroup":               "acls",
	"BasicAuth":              "basicauth_credentials",
	"CACertificate":          "ca_certificates",
	"Certificate":            "certificates",
	"Consumer":               "consumers",
	"ConsumerGroup":          "consumer_groups",
	"FilterChain":            "filter_chains",
	"HMACAuth":               "hmacauth_credentials",
	"JWTAuth":                "jwt_secrets",
	"License":                "licenses",
	"Key":                    "keys",
	"KeyAuth":                "keyauth_credentials",
	"KeySet":                 "keysets",
	"MTLSAuth":               "mtls_auth_credentials",
	"Oauth2Cred":             "oauth2_credentials",
	"Partial":                "partials",
	"Plugin":                 "plugins",
	"RBACEndpointPermission": "rbac-endpointpermission",
	"RBACRole":               "rbac-role",
	"Route":                  "routes",
	"SNI":                    "snis",
	"Service":                "services",
	"Target":                 "targets",
	"Upstream":               "upstreams",
	"Vault":                  "vaults",
}

// Some entities in Konnect have different names compared to Kong Gateway
var kongToKonnectEntitiesMap = map[string]string{
	"services":    "service",
	"routes":      "route",
	"upstreams":   "upstream",
	"targets":     "target",
	"jwt_secrets": "jwt",
}

func (s *Sanitizer) fetchEntitySchema(entityName string, field reflect.Value) (string, kong.Schema, error) {
	var (
		schema kong.Schema
		err    error
	)

	// for unit tests, we may have an uninitialized client
	if s.client == nil {
		return entityName, nil, fmt.Errorf("kong client is not initialized")
	}

	if entityName == Plugin {
		pluginName := field.FieldByName("Name").String()
		schema, err = s.pluginSchemasCache.Get(s.ctx, pluginName)
		return pluginName, schema, err
	}

	if entityName == Partial {
		partialType := field.FieldByName("Type").String()
		schema, err = s.partialSchemasCache.Get(s.ctx, partialType)
		return partialType, schema, err
	}

	// important so that entities like FPlugin, FService, etc. are not tried for schema fetching
	entityName, ok := entityMap[entityName]
	if !ok {
		return "", nil, fmt.Errorf("unknown entity type %s", entityName)
	}

	if s.isKonnect {
		schema, err = s.getKonnectEntitySchema(entityName)
		return entityName, schema, err
	}

	schema, err = s.client.Schemas.Get(s.ctx, entityName)
	if err != nil {
		return entityName, nil, err
	}
	if schema == nil {
		return entityName, nil, fmt.Errorf("schema for entity type %s not found", entityName)
	}
	return entityName, schema, nil
}

func (s *Sanitizer) getKonnectEntitySchema(entityType string) (kong.Schema, error) {
	var (
		schema map[string]interface{}
		ok     bool
	)

	entityType, ok = kongToKonnectEntitiesMap[entityType]
	if !ok {
		return schema, nil
	}

	endpoint := fmt.Sprintf("/v1/schemas/json/%s", entityType)

	req, err := s.client.NewRequest(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return schema, err
	}
	resp, err := s.client.Do(s.ctx, req, &schema)
	if resp == nil {
		return schema, fmt.Errorf("invalid HTTP response: %w", err)
	}

	return schema, nil
}
