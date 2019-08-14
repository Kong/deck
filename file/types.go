package file

import "github.com/hbagdi/go-kong/kong"

// Service represents a Kong Service and it's associated routes and plugins.
type Service struct {
	kong.Service `yaml:",inline,omitempty"`
	Routes       []*Route  `yaml:",omitempty"`
	Plugins      []*Plugin `yaml:",omitempty"`
}

// Route represents a Kong Route and it's associated plugins.
type Route struct {
	kong.Route `yaml:",inline,omitempty"`
	Plugins    []*Plugin `yaml:",omitempty"`
}

// Upstream represents a Kong Upstream and it's associated targets.
type Upstream struct {
	kong.Upstream `yaml:",inline,omitempty"`
	Targets       []*Target `yaml:",omitempty"`
}

// Target represents a Kong Target.
type Target struct {
	kong.Target `yaml:",inline,omitempty"`
}

// Certificate represents a Kong Certificate.
type Certificate struct {
	kong.Certificate `yaml:",inline,omitempty"`
}

// Plugin represents a plugin in Kong.
type Plugin struct {
	kong.Plugin `yaml:",inline,omitempty"`
}

// Consumer represents a consumer in Kong.
type Consumer struct {
	kong.Consumer `yaml:",inline,omitempty"`
	Plugins       []*Plugin        `yaml:",omitempty"`
	KeyAuths      []*kong.KeyAuth  `yaml:"keyauth_credentials,omitempty"`
	HMACAuths     []*kong.HMACAuth `yaml:"hmacauth_credentials,omitempty"`
	JWTAuths      []*kong.JWTAuth  `yaml:"jwt_secrets,omitempty"`
}

// Info contains meta-data of the file.
type Info struct {
	SelectorTags []string `yaml:"select_tags,omitempty"`
}

// Content represents a serialized Kong state.
type Content struct {
	FormatVersion string        `yaml:"_format_version,omitempty"`
	Info          Info          `yaml:"_info,omitempty"`
	Services      []Service     `yaml:",omitempty"`
	Upstreams     []Upstream    `yaml:",omitempty"`
	Certificates  []Certificate `yaml:",omitempty"`
	Plugins       []Plugin      `yaml:",omitempty"`
	Consumers     []Consumer    `yaml:",omitempty"`
}
