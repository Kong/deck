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
	"Route":               {"Methods": {}},

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
	"dictionary_name": {}, // present in rla, upstream_oauth
}

// dynamically generated maps of exempted fields from schemas
var entityLevelExemptedFieldsFromSchema = map[string]map[string]bool{}
