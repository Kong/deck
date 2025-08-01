package sanitize

// Top-level content fields that should not be sanitized
var topLevelExemptedFields = map[string]struct{}{
	"Info":          {},
	"FormatVersion": {},
	"Konnect":       {},
	"Transform":     {},
	"Workspace":     {},
}
