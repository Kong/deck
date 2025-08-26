package sanitize

// Top-level content fields that should not be sanitized
var topLevelExemptedFields = map[string]struct{}{
	"Info":          {},
	"FormatVersion": {},
	"Konnect":       {},
	"Transform":     {},
	"Workspace":     {},
}

// Static fields that should be skipped for specific entities
// These fields either require special handling (like Keys)
// or need not be sanitized like plugin.name
var entityLevelExemptedFields = map[string]map[string]struct{}{
	// Entity level exemptions
	"ConsumerGroupPlugin": {"Name": {}},
	"Partial":             {"Type": {}},
	"PartialLink":         {"Path": {}},
	"Plugin":              {"Name": {}},
	"Route":               {"Methods": {}, "Expression": {}},

	// Special handling
	"CACertificate": {"Cert": {}, "CertDigest": {}},
	"FCertificate":  {"Cert": {}, "Key": {}},
	"Key":           {"PEM": {}, "JWK": {}, "KID": {}},
	"Vault":         {"Name": {}, "Prefix": {}},
}

// Config-level fields that should not be sanitized
var configLevelExemptedFields = map[string]struct{}{
	"ID": {},

	// Plugin specific exemptions, that can't be generated from schema
	"dictionary_name":          {}, // present in rla, upstream_oauth
	"lock_dictionary_name":     {}, // present in rla
	"access":                   {}, // present in pre and post-function
	"request_jq_program":       {}, // present in jq
	"response_jq_program":      {}, // present in jq
	"log_by_lua":               {},
	"shm_name":                 {}, // present in acme
	"schema_version":           {}, // present in confluent
	"message_by_lua_functions": {}, // present in confluent
	"custom_fields_by_lua":     {}, // present in http-log, tcp-log
	"api_spec":                 {}, // present in oas-validation
	"anthropic_version":        {}, // present in ai-request-transformer
	"schema":                   {}, // present in request-validator
	"by_lua":                   {}, // present in request-callout

	// Vault specific exemptions, that can't be generated from schema
	"prefix": {}, // present in env
}

var entitiesToHandleDifferently = map[string]struct{}{
	"CACertificate": {},
	"FCertificate":  {},
	"Key":           {},
	"Route":         {},
}

// dynamically generated maps of exempted fields from schemas
var entityLevelExemptedFieldsFromSchema = map[string]map[string]bool{}
