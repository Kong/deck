package file

import (
	"encoding/json"

	"github.com/hbagdi/go-kong/kong"
)

// Format is a file format for Kong's configuration.
type Format string

type id interface {
	id() string
}

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

// id is used for sorting.
func (s Service) id() string {
	if s.ID != nil {
		return *s.ID
	}
	if s.Name != nil {
		return *s.Name
	}
	return ""
}

// Route represents a Kong Route and it's associated plugins.
type Route struct {
	kong.Route `yaml:",inline,omitempty"`
	Plugins    []*Plugin `json:"plugins,omitempty" yaml:",omitempty"`
}

// id is used for sorting.
func (r Route) id() string {
	if r.ID != nil {
		return *r.ID
	}
	if r.Name != nil {
		return *r.Name
	}
	return ""
}

// Upstream represents a Kong Upstream and it's associated targets.
type Upstream struct {
	kong.Upstream `yaml:",inline,omitempty"`
	Targets       []*Target `json:"targets,omitempty" yaml:",omitempty"`
}

// id is used for sorting.
func (u Upstream) id() string {
	if u.ID != nil {
		return *u.ID
	}
	if u.Name != nil {
		return *u.Name
	}
	return ""
}

// Target represents a Kong Target.
type Target struct {
	kong.Target `yaml:",inline,omitempty"`
}

// id is used for sorting.
func (t Target) id() string {
	if t.ID != nil {
		return *t.ID
	}
	if t.Target.Target != nil {
		return *t.Target.Target
	}
	return ""
}

// Certificate represents a Kong Certificate.
type Certificate struct {
	kong.Certificate `yaml:",inline,omitempty"`
}

// id is used for sorting.
func (c Certificate) id() string {
	if c.ID != nil {
		return *c.ID
	}
	if c.Cert != nil {
		return *c.Cert
	}
	return ""
}

// CACertificate represents a Kong CACertificate.
type CACertificate struct {
	kong.CACertificate `yaml:",inline,omitempty"`
}

// id is used for sorting.
func (c CACertificate) id() string {
	if c.ID != nil {
		return *c.ID
	}
	if c.Cert != nil {
		return *c.Cert
	}
	return ""
}

// Plugin represents a plugin in Kong.
type Plugin struct {
	kong.Plugin `yaml:",inline,omitempty"`
}

// foo is a shadow type of Plugin.
// It is used for custom marshalling of plugin.
type foo struct {
	CreatedAt *int               `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	ID        *string            `json:"id,omitempty" yaml:"id,omitempty"`
	Name      *string            `json:"name,omitempty" yaml:"name,omitempty"`
	Config    kong.Configuration `json:"config,omitempty" yaml:"config,omitempty"`
	Service   string             `json:"service,omitempty" yaml:",omitempty"`
	Consumer  string             `json:"consumer,omitempty" yaml:",omitempty"`
	Route     string             `json:"route,omitempty" yaml:",omitempty"`
	Enabled   *bool              `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	RunOn     *string            `json:"run_on,omitempty" yaml:"run_on,omitempty"`
	Protocols []*string          `json:"protocols,omitempty" yaml:"protocols,omitempty"`
	Tags      []*string          `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func copyToFoo(p Plugin) foo {
	f := foo{}
	if p.ID != nil {
		f.ID = p.ID
	}
	if p.Name != nil {
		f.Name = p.Name
	}
	if p.Enabled != nil {
		f.Enabled = p.Enabled
	}
	if p.RunOn != nil {
		f.RunOn = p.RunOn
	}
	if p.Protocols != nil {
		f.Protocols = p.Protocols
	}
	if p.Tags != nil {
		f.Tags = p.Tags
	}
	if p.Config != nil {
		f.Config = p.Config
	}
	if p.Plugin.Consumer != nil {
		f.Consumer = *p.Plugin.Consumer.ID
	}
	if p.Plugin.Route != nil {
		f.Route = *p.Plugin.Route.ID
	}
	if p.Plugin.Service != nil {
		f.Service = *p.Plugin.Service.ID
	}
	return f
}

// MarshalYAML is a custom marshal method to handle
// foreign references.
func (p Plugin) MarshalYAML() (interface{}, error) {
	return copyToFoo(p), nil
}

// MarshalJSON is a custom marshal method to handle
// foreign references.
func (p Plugin) MarshalJSON() ([]byte, error) {
	f := copyToFoo(p)
	return json.Marshal(f)
}

// id is used for sorting.
func (p Plugin) id() string {
	if p.ID != nil {
		return *p.ID
	}
	// concat plugin name and foreign relations
	key := ""
	key = *p.Name
	if p.Consumer != nil {
		key += *p.Consumer.ID
	}
	if p.Route != nil {
		key += *p.Route.ID
	}
	if p.Service != nil {
		key += *p.Service.ID
	}
	return key
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

// id is used for sorting.
func (c Consumer) id() string {
	if c.ID != nil {
		return *c.ID
	}
	if c.Username != nil {
		return *c.Username
	}
	return ""
}

// Info contains meta-data of the file.
type Info struct {
	SelectorTags []string `json:"select_tags,omitempty" yaml:"select_tags,omitempty"`
}

// Content represents a serialized Kong state.
type Content struct {
	FormatVersion string `json:"_format_version,omitempty" yaml:"_format_version,omitempty"`
	Info          *Info  `json:"_info,omitempty" yaml:"_info,omitempty"`
	Workspace     string `json:"_workspace,omitempty" yaml:"_workspace,omitempty"`

	Services       []Service       `json:"services,omitempty" yaml:",omitempty"`
	Routes         []Route         `json:"routes,omitempty" yaml:",omitempty"`
	Upstreams      []Upstream      `json:"upstreams,omitempty" yaml:",omitempty"`
	Certificates   []Certificate   `json:"certificates,omitempty" yaml:",omitempty"`
	CACertificates []CACertificate `json:"ca_certificates,omitempty" yaml:"ca_certificates,omitempty"`
	Plugins        []Plugin        `json:"plugins,omitempty" yaml:",omitempty"`
	Consumers      []Consumer      `json:"consumers,omitempty" yaml:",omitempty"`
}
