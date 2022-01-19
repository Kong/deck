package konnect

const (
	// KonnectManagedPluginTag is used by Konnect to tag internally-managed plugins
	KonnectManagedPluginTag = "konnect-managed-plugin"
)

func emptyString(p *string) bool {
	return p == nil || *p == ""
}
