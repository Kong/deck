package file

import "github.com/hbagdi/go-kong/kong"

// Format is a file format for Kong's configuration.
type Format string

const (
	// JSON is JSON file format.
	JSON = "JSON"
	// YAML if YAML file format.
	YAML = "YAML"
)

// Service represents a Kong Service and it's associated routes and plugins.
type Service struct {
	kong.Service `yaml:",inline,omitempty"`
	Routes       []*Route  `json:"routes,omitempty" yaml:",omitempty"`
	Plugins      []*Plugin `json:"plugins,omitempty" yaml:",omitempty"`
}

// Route represents a Kong Route and it's associated plugins.
type Route struct {
	kong.Route `yaml:",inline,omitempty"`
	Plugins    []*Plugin `json:"plugins,omitempty" yaml:",omitempty"`
}

// Upstream represents a Kong Upstream and it's associated targets.
type Upstream struct {
	kong.Upstream `yaml:",inline,omitempty"`
	Targets       []*Target `json:"targets,omitempty" yaml:",omitempty"`
}

// Target represents a Kong Target.
type Target struct {
	kong.Target `yaml:",inline,omitempty"`
}

// Certificate represents a Kong Certificate.
type Certificate struct {
	kong.Certificate `yaml:",inline,omitempty"`
}

// CACertificate represents a Kong CACertificate.
type CACertificate struct {
	kong.CACertificate `yaml:",inline,omitempty"`
}

// Plugin represents a plugin in Kong.
type Plugin struct {
	kong.Plugin `yaml:",inline,omitempty"`
}

// Consumer represents a consumer in Kong.
type Consumer struct {
	kong.Consumer `yaml:",inline,omitempty"`
	Plugins       []*Plugin                `json:"plugins,omitempty" yaml:",omitempty"`
	KeyAuths      []*kong.KeyAuth          `json:"keyauth_credentials,omitempty" yaml:"keyauth_credentials,omitempty"`
	HMACAuths     []*kong.HMACAuth         `json:"hmacauth_credentials,omitempty" yaml:"hmacauth_credentials,omitempty"`
	JWTAuths      []*kong.JWTAuth          `json:"jwt_secrets,omitempty" yaml:"jwt_secrets,omitempty"`
	BasicAuths    []*kong.BasicAuth        `json:"basicauth_credentials,omitempty" yaml:"basicauth_credentials,omitempty"`
	Oauth2Creds   []*kong.Oauth2Credential `json:"oauth2_credentials,omitempty" yaml:"oauth2_credentials,omitempty"`
	ACLGroups     []*kong.ACLGroup         `json:"acls,omitempty" yaml:"acls,omitempty"`
}

// Info contains meta-data of the file.
type Info struct {
	SelectorTags []string `json:"select_tags,omitempty" yaml:"select_tags,omitempty"`
}

// Content represents a serialized Kong state.
type Content struct {
	FormatVersion  string          `json:"_format_version,omitempty" yaml:"_format_version,omitempty"`
	Info           *Info           `json:"_info,omitempty" yaml:"_info,omitempty"`
	Workspace      string          `json:"_workspace,omitempty" yaml:"_workspace,omitempty"`
	Services       []Service       `json:"services,omitempty" yaml:",omitempty"`
	Upstreams      []Upstream      `json:"upstreams,omitempty" yaml:",omitempty"`
	Certificates   []Certificate   `json:"certificates,omitempty" yaml:",omitempty"`
	CACertificates []CACertificate `json:"ca_certificates,omitempty" yaml:"ca_certificates,omitempty"`
	Plugins        []Plugin        `json:"plugins,omitempty" yaml:",omitempty"`
	Consumers      []Consumer      `json:"consumers,omitempty" yaml:",omitempty"`
}
