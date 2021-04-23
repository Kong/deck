package file

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/pkg/errors"
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

// FService represents a Kong Service and it's associated routes and plugins.
// +k8s:deepcopy-gen=true
type FService struct {
	kong.Service
	Routes  []*FRoute  `json:"routes,omitempty" yaml:",omitempty"`
	Plugins []*FPlugin `json:"plugins,omitempty" yaml:",omitempty"`

	// sugar property
	URL *string `json:"url,omitempty" yaml:",omitempty"`
}

// id is used for sorting.
func (s FService) id() string {
	if s.ID != nil {
		return *s.ID
	}
	if s.Name != nil {
		return *s.Name
	}
	return ""
}

type service struct {
	ClientCertificate *string    `json:"client_certificate,omitempty" yaml:"client_certificate,omitempty"`
	ConnectTimeout    *int       `json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"`
	CreatedAt         *int       `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	Host              *string    `json:"host,omitempty" yaml:"host,omitempty"`
	ID                *string    `json:"id,omitempty" yaml:"id,omitempty"`
	Name              *string    `json:"name,omitempty" yaml:"name,omitempty"`
	Path              *string    `json:"path,omitempty" yaml:"path,omitempty"`
	Port              *int       `json:"port,omitempty" yaml:"port,omitempty"`
	Protocol          *string    `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	ReadTimeout       *int       `json:"read_timeout,omitempty" yaml:"read_timeout,omitempty"`
	Retries           *int       `json:"retries,omitempty" yaml:"retries,omitempty"`
	UpdatedAt         *int       `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	WriteTimeout      *int       `json:"write_timeout,omitempty" yaml:"write_timeout,omitempty"`
	Tags              []*string  `json:"tags,omitempty" yaml:"tags,omitempty"`
	TLSVerify         *bool      `json:"tls_verify,omitempty" yaml:"tls_verify,omitempty"`
	TLSVerifyDepth    *int       `json:"tls_verify_depth,omitempty" yaml:"tls_verify_depth,omitempty"`
	CACertificates    []*string  `json:"ca_certificates,omitempty" yaml:"ca_certificates,omitempty"`
	Routes            []*FRoute  `json:"routes,omitempty" yaml:",omitempty"`
	Plugins           []*FPlugin `json:"plugins,omitempty" yaml:",omitempty"`

	// sugar property
	URL *string `json:"url,omitempty" yaml:",omitempty"`
}

func copyToService(fService FService) service {
	s := service{}
	if fService.ClientCertificate != nil &&
		!utils.Empty(fService.ClientCertificate.ID) {
		s.ClientCertificate = kong.String(*fService.ClientCertificate.ID)
	}
	s.CACertificates = fService.CACertificates
	s.TLSVerify = fService.TLSVerify
	s.TLSVerifyDepth = fService.TLSVerifyDepth
	s.ConnectTimeout = fService.ConnectTimeout
	s.CreatedAt = fService.CreatedAt
	s.Host = fService.Host
	s.ID = fService.ID
	s.Name = fService.Name
	s.Path = fService.Path
	s.Port = fService.Port
	s.Protocol = fService.Protocol
	s.ReadTimeout = fService.ReadTimeout
	s.Retries = fService.Retries
	s.UpdatedAt = fService.UpdatedAt
	s.WriteTimeout = fService.WriteTimeout
	s.Tags = fService.Tags
	s.Routes = fService.Routes
	s.Plugins = fService.Plugins

	return s
}

func unwrapURL(urlString string, fService *FService) error {
	parsed, err := url.Parse(urlString)
	if err != nil {
		return errors.New("invaid url: " + urlString)
	}
	if parsed.Scheme == "" {
		return errors.New("invalid url:" + urlString)
	}

	fService.Protocol = kong.String(parsed.Scheme)

	fService.Port = kong.Int(80)
	if parsed.Scheme == "https" {
		fService.Port = kong.Int(443)
	}

	if parsed.Host != "" {
		hostPort := strings.Split(parsed.Host, ":")
		fService.Host = kong.String(hostPort[0])
		if len(hostPort) > 1 {
			port, err := strconv.Atoi(hostPort[1])
			if err == nil {
				fService.Port = kong.Int(port)
			}
		}
	}
	if parsed.Path != "" {
		fService.Path = kong.String(parsed.Path)
	}
	return nil
}

func copyFromService(service service, fService *FService) error {
	if service.ClientCertificate != nil &&
		!utils.Empty(service.ClientCertificate) {
		fService.ClientCertificate = &kong.Certificate{
			ID: kong.String(*service.ClientCertificate),
		}
	}
	if !utils.Empty(service.URL) {
		err := unwrapURL(*service.URL, fService)
		if err != nil {
			return err
		}
	}
	fService.ConnectTimeout = service.ConnectTimeout
	fService.CreatedAt = service.CreatedAt
	fService.ID = service.ID
	fService.Name = service.Name
	if service.Protocol != nil {
		fService.Protocol = service.Protocol
	}
	if service.Host != nil {
		fService.Host = service.Host
	}
	if service.Port != nil {
		fService.Port = service.Port
	}
	if service.Path != nil {
		fService.Path = service.Path
	}
	fService.ReadTimeout = service.ReadTimeout
	fService.Retries = service.Retries
	fService.UpdatedAt = service.UpdatedAt
	fService.WriteTimeout = service.WriteTimeout
	fService.Tags = service.Tags
	fService.CACertificates = service.CACertificates
	fService.TLSVerify = service.TLSVerify
	fService.TLSVerifyDepth = service.TLSVerifyDepth
	fService.Routes = service.Routes
	fService.Plugins = service.Plugins
	return nil
}

// MarshalYAML is a custom marshal to handle
// SNI.
func (s FService) MarshalYAML() (interface{}, error) {
	return copyToService(s), nil
}

// UnmarshalYAML is a custom marshal method to handle
// foreign references.
func (s *FService) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var service service
	if err := unmarshal(&service); err != nil {
		return err
	}
	return copyFromService(service, s)
}

// MarshalJSON is a custom marshal method to handle
// foreign references.
func (s FService) MarshalJSON() ([]byte, error) {
	service := copyToService(s)
	return json.Marshal(service)
}

// UnmarshalJSON is a custom marshal method to handle
// foreign references.
func (s *FService) UnmarshalJSON(b []byte) error {
	var service service
	err := json.Unmarshal(b, &service)
	if err != nil {
		return err
	}
	return copyFromService(service, s)
}

// FRoute represents a Kong Route and it's associated plugins.
// +k8s:deepcopy-gen=true
type FRoute struct {
	kong.Route `yaml:",inline,omitempty"`
	Plugins    []*FPlugin `json:"plugins,omitempty" yaml:",omitempty"`
}

// id is used for sorting.
func (r FRoute) id() string {
	if r.ID != nil {
		return *r.ID
	}
	if r.Name != nil {
		return *r.Name
	}
	return ""
}

// FUpstream represents a Kong Upstream and it's associated targets.
// +k8s:deepcopy-gen=true
type FUpstream struct {
	kong.Upstream `yaml:",inline,omitempty"`
	Targets       []*FTarget `json:"targets,omitempty" yaml:",omitempty"`
}

// id is used for sorting.
func (u FUpstream) id() string {
	if u.ID != nil {
		return *u.ID
	}
	if u.Name != nil {
		return *u.Name
	}
	return ""
}

// FTarget represents a Kong Target.
// +k8s:deepcopy-gen=true
type FTarget struct {
	kong.Target `yaml:",inline,omitempty"`
}

// id is used for sorting.
func (t FTarget) id() string {
	if t.ID != nil {
		return *t.ID
	}
	if t.Target.Target != nil {
		return *t.Target.Target
	}
	return ""
}

// FCertificate represents a Kong Certificate.
// +k8s:deepcopy-gen=true
type FCertificate struct {
	ID        *string    `json:"id,omitempty" yaml:"id,omitempty"`
	Cert      *string    `json:"cert,omitempty" yaml:"cert,omitempty"`
	Key       *string    `json:"key,omitempty" yaml:"key,omitempty"`
	CreatedAt *int64     `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	Tags      []*string  `json:"tags,omitempty" yaml:"tags,omitempty"`
	SNIs      []kong.SNI `json:"snis,omitempty" yaml:"snis,omitempty"`
}

// id is used for sorting.
func (c FCertificate) id() string {
	if c.ID != nil {
		return *c.ID
	}
	if c.Cert != nil {
		return *c.Cert
	}
	return ""
}

// FCACertificate represents a Kong CACertificate.
// +k8s:deepcopy-gen=true
type FCACertificate struct {
	kong.CACertificate `yaml:",inline,omitempty"`
}

// id is used for sorting.
func (c FCACertificate) id() string {
	if c.ID != nil {
		return *c.ID
	}
	if c.Cert != nil {
		return *c.Cert
	}
	return ""
}

// FPlugin represents a plugin in Kong.
// +k8s:deepcopy-gen=true
type FPlugin struct {
	kong.Plugin `yaml:",inline,omitempty"`

	ConfigSource *string `json:"_config,omitempty" yaml:"_config,omitempty"`
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

	ConfigSource *string `json:"_config,omitempty" yaml:"_config,omitempty"`
}

func copyToFoo(p FPlugin) foo {
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
	if p.ConfigSource != nil {
		f.ConfigSource = p.ConfigSource
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

func copyFromFoo(f foo, p *FPlugin) {
	if f.ID != nil {
		p.ID = f.ID
	}
	if f.Name != nil {
		p.Name = f.Name
	}
	if f.Enabled != nil {
		p.Enabled = f.Enabled
	}
	if f.RunOn != nil {
		p.RunOn = f.RunOn
	}
	if f.Protocols != nil {
		p.Protocols = f.Protocols
	}
	if f.Tags != nil {
		p.Tags = f.Tags
	}
	if f.Config != nil {
		p.Config = f.Config
	}
	if f.ConfigSource != nil {
		p.ConfigSource = f.ConfigSource
	}
	if f.Consumer != "" {
		p.Consumer = &kong.Consumer{
			ID: kong.String(f.Consumer),
		}
	}
	if f.Route != "" {
		p.Route = &kong.Route{
			ID: kong.String(f.Route),
		}
	}
	if f.Service != "" {
		p.Service = &kong.Service{
			ID: kong.String(f.Service),
		}
	}
}

// MarshalYAML is a custom marshal method to handle
// foreign references.
func (p FPlugin) MarshalYAML() (interface{}, error) {
	return copyToFoo(p), nil
}

// UnmarshalYAML is a custom marshal method to handle
// foreign references.
func (p *FPlugin) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var f foo
	if err := unmarshal(&f); err != nil {
		return err
	}
	copyFromFoo(f, p)
	return nil
}

// MarshalJSON is a custom marshal method to handle
// foreign references.
func (p FPlugin) MarshalJSON() ([]byte, error) {
	f := copyToFoo(p)
	return json.Marshal(f)
}

// UnmarshalJSON is a custom marshal method to handle
// foreign references.
func (p *FPlugin) UnmarshalJSON(b []byte) error {
	var f foo
	err := json.Unmarshal(b, &f)
	if err != nil {
		return err
	}
	copyFromFoo(f, p)
	return nil
}

// id is used for sorting.
func (p FPlugin) id() string {
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

// FConsumer represents a consumer in Kong.
// +k8s:deepcopy-gen=true
type FConsumer struct {
	kong.Consumer `yaml:",inline,omitempty"`
	Plugins       []*FPlugin               `json:"plugins,omitempty" yaml:",omitempty"`
	KeyAuths      []*kong.KeyAuth          `json:"keyauth_credentials,omitempty" yaml:"keyauth_credentials,omitempty"`
	HMACAuths     []*kong.HMACAuth         `json:"hmacauth_credentials,omitempty" yaml:"hmacauth_credentials,omitempty"`
	JWTAuths      []*kong.JWTAuth          `json:"jwt_secrets,omitempty" yaml:"jwt_secrets,omitempty"`
	BasicAuths    []*kong.BasicAuth        `json:"basicauth_credentials,omitempty" yaml:"basicauth_credentials,omitempty"`
	Oauth2Creds   []*kong.Oauth2Credential `json:"oauth2_credentials,omitempty" yaml:"oauth2_credentials,omitempty"`
	ACLGroups     []*kong.ACLGroup         `json:"acls,omitempty" yaml:"acls,omitempty"`
	MTLSAuths     []*kong.MTLSAuth         `json:"mtls_auth_credentials,omitempty" yaml:"mtls_auth_credentials,omitempty"`
}

// id is used for sorting.
func (c FConsumer) id() string {
	if c.ID != nil {
		return *c.ID
	}
	if c.Username != nil {
		return *c.Username
	}
	return ""
}

// FRBACRole represents an RBACRole in Kong
// +k8s:deepcopy-gen=true
type FRBACRole struct {
	kong.RBACRole       `yaml:",inline,omitempty"`
	EndpointPermissions []*FRBACEndpointPermission `json:"endpoint_permissions,omitempty" yaml:"endpoint_permissions,omitempty"` //nolint
}

// +k8s:deepcopy-gen=true
type FRBACEndpointPermission struct {
	kong.RBACEndpointPermission `yaml:",inline,omitempty"`
}

// Info contains meta-data of the file.
// +k8s:deepcopy-gen=true
type Info struct {
	SelectorTags []string `json:"select_tags,omitempty" yaml:"select_tags,omitempty"`
}

// +k8s:deepcopy-gen=true
type Kong struct {
	Service *FService `json:"service,omitempty" yaml:"service,omitempty"`
}

// +k8s:deepcopy-gen=true
type Implementation struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	Kong *Kong  `json:"kong,omitempty" yaml:"kong,omitempty"`
}

// +k8s:deepcopy-gen=true
type FServiceVersion struct {
	ID             *string         `json:"id,omitempty" yaml:"id,omitempty"`
	Version        *string         `json:"version,omitempty" yaml:"version,omitempty"`
	Implementation *Implementation `json:"implementation,omitempty" yaml:"implementation,omitempty"`
}

// +k8s:deepcopy-gen=true
type FServicePackage struct {
	ID          *string           `json:"id,omitempty" yaml:"id,omitempty"`
	Name        *string           `json:"name,omitempty" yaml:"name,omitempty"`
	Description *string           `json:"description,omitempty" yaml:"description,omitempty"`
	Versions    []FServiceVersion `json:"versions,omitempty" yaml:"versions,omitempty"`
}

// id is used for sorting.
func (s FServicePackage) id() string {
	if s.ID != nil {
		return *s.ID
	}
	if s.Name != nil {
		return *s.Name
	}
	return ""
}

//go:generate go run ./codegen/main.go

// Content represents a serialized Kong state.
// +k8s:deepcopy-gen=true
type Content struct {
	FormatVersion string `json:"_format_version,omitempty" yaml:"_format_version,omitempty"`
	Info          *Info  `json:"_info,omitempty" yaml:"_info,omitempty"`
	Workspace     string `json:"_workspace,omitempty" yaml:"_workspace,omitempty"`

	Services       []FService       `json:"services,omitempty" yaml:",omitempty"`
	Routes         []FRoute         `json:"routes,omitempty" yaml:",omitempty"`
	Consumers      []FConsumer      `json:"consumers,omitempty" yaml:",omitempty"`
	Plugins        []FPlugin        `json:"plugins,omitempty" yaml:",omitempty"`
	Upstreams      []FUpstream      `json:"upstreams,omitempty" yaml:",omitempty"`
	Certificates   []FCertificate   `json:"certificates,omitempty" yaml:",omitempty"`
	CACertificates []FCACertificate `json:"ca_certificates,omitempty" yaml:"ca_certificates,omitempty"`

	RBACRoles []FRBACRole `json:"rbac_roles,omitempty" yaml:"rbac_roles,omitempty"`

	PluginConfigs map[string]kong.Configuration `json:"_plugin_configs,omitempty" yaml:"_plugin_configs,omitempty"`

	ServicePackages []FServicePackage `json:"service_packages,omitempty" yaml:"service_packages,omitempty"`
}
