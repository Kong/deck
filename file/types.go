package file

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

// Format is a file format for Kong's configuration.
type Format string

type sortable interface {
	sortKey() string
}

const (
	// JSON is JSON file format.
	JSON = "JSON"
	// YAML if YAML file format.
	YAML = "YAML"
)

const (
	httpPort  = 80
	httpsPort = 443
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

// sortKey is used for sorting.
func (s FService) sortKey() string {
	if s.Name != nil {
		return *s.Name
	}
	if s.ID != nil {
		return *s.ID
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
		return fmt.Errorf("invalid url: " + urlString)
	}
	if parsed.Scheme == "" {
		return fmt.Errorf("invalid url:" + urlString)
	}

	fService.Protocol = kong.String(parsed.Scheme)

	fService.Port = kong.Int(httpPort)
	if parsed.Scheme == "https" {
		fService.Port = kong.Int(httpsPort)
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

// sortKey is used for sorting.
func (r FRoute) sortKey() string {
	if r.Name != nil {
		return *r.Name
	}
	if r.ID != nil {
		return *r.ID
	}
	return ""
}

// FUpstream represents a Kong Upstream and it's associated targets.
// +k8s:deepcopy-gen=true
type FUpstream struct {
	kong.Upstream `yaml:",inline,omitempty"`
	Targets       []*FTarget `json:"targets,omitempty" yaml:",omitempty"`
}

// sortKey is used for sorting.
func (u FUpstream) sortKey() string {
	if u.Name != nil {
		return *u.Name
	}
	if u.ID != nil {
		return *u.ID
	}
	return ""
}

// FTarget represents a Kong Target.
// +k8s:deepcopy-gen=true
type FTarget struct {
	kong.Target `yaml:",inline,omitempty"`
}

// sortKey is used for sorting.
func (t FTarget) sortKey() string {
	if t.Target.Target != nil {
		return *t.Target.Target
	}
	if t.ID != nil {
		return *t.ID
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

// sortKey is used for sorting.
func (c FCertificate) sortKey() string {
	if c.Cert != nil {
		return *c.Cert
	}
	if c.ID != nil {
		return *c.ID
	}
	return ""
}

// FCACertificate represents a Kong CACertificate.
// +k8s:deepcopy-gen=true
type FCACertificate struct {
	kong.CACertificate `yaml:",inline,omitempty"`
}

// sortKey is used for sorting.
func (c FCACertificate) sortKey() string {
	if c.Cert != nil {
		return *c.Cert
	}
	if c.ID != nil {
		return *c.ID
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

// sortKey is used for sorting.
func (p FPlugin) sortKey() string {
	// concat plugin name and foreign relations
	if p.Name != nil {
		key := *p.Name
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
	if p.ID != nil {
		return *p.ID
	}
	return ""
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

// sortKey is used for sorting.
func (c FConsumer) sortKey() string {
	if c.Username != nil {
		return *c.Username
	}
	if c.ID != nil {
		return *c.ID
	}
	return ""
}

// FRBACRole represents an RBACRole in Kong
// +k8s:deepcopy-gen=true
type FRBACRole struct {
	kong.RBACRole       `yaml:",inline,omitempty"`
	EndpointPermissions []*FRBACEndpointPermission `json:"endpoint_permissions,omitempty" yaml:"endpoint_permissions,omitempty"` //nolint
}

// FRBACEndpointPermission is a wrapper type for RBACEndpointPermission.
// +k8s:deepcopy-gen=true
type FRBACEndpointPermission struct {
	kong.RBACEndpointPermission `yaml:",inline,omitempty"`
}

// KongDefaults represents default values that are filled in
// for entities with corresponding missing properties.
// +k8s:deepcopy-gen=true
type KongDefaults struct {
	Service  *kong.Service  `json:"service,omitempty" yaml:"service,omitempty"`
	Route    *kong.Route    `json:"route,omitempty" yaml:"route,omitempty"`
	Upstream *kong.Upstream `json:"upstream,omitempty" yaml:"upstream,omitempty"`
	Target   *kong.Target   `json:"target,omitempty" yaml:"target,omitempty"`
}

// Info contains meta-data of the file.
// +k8s:deepcopy-gen=true
type Info struct {
	SelectorTags []string     `json:"select_tags,omitempty" yaml:"select_tags,omitempty"`
	Defaults     KongDefaults `json:"defaults,omitempty" yaml:"defaults,omitempty"`
}

// Kong represents Kong implementation of a Service in Konnect.
// +k8s:deepcopy-gen=true
type Kong struct {
	Service *FService `json:"service,omitempty" yaml:"service,omitempty"`
}

// Implementation represents an implementation of a Service version in Konnect.
// +k8s:deepcopy-gen=true
type Implementation struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	Kong *Kong  `json:"kong,omitempty" yaml:"kong,omitempty"`
}

// FServiceVersion represents a Service version in Konnect.
// The type is duplicated because only a single document is
// exported in file while the API allows for multiple documents.
// +k8s:deepcopy-gen=true
type FServiceVersion struct {
	ID             *string         `json:"id,omitempty" yaml:"id,omitempty"`
	Version        *string         `json:"version,omitempty" yaml:"version,omitempty"`
	Implementation *Implementation `json:"implementation,omitempty" yaml:"implementation,omitempty"`
	Document       *FDocument      `json:"document,omitempty" yaml:"document,omitempty"`
}

// FServicePackage represents a Service package and its associated entities.
// +k8s:deepcopy-gen=true
type FServicePackage struct {
	ID          *string           `json:"id,omitempty" yaml:"id,omitempty"`
	Name        *string           `json:"name,omitempty" yaml:"name,omitempty"`
	Description *string           `json:"description,omitempty" yaml:"description,omitempty"`
	Versions    []FServiceVersion `json:"versions,omitempty" yaml:"versions,omitempty"`
	Document    *FDocument        `json:"document,omitempty" yaml:"document,omitempty"`
}

// FDocument represents a document in Konnect.
// The type has been duplicated because the documents are altered
// before they are exported to the state file
// for better user experience.
// +k8s:deepcopy-gen=true
type FDocument struct {
	ID        *string `json:"id,omitempty" yaml:"id,omitempty"`
	Path      *string `json:"path,omitempty" yaml:"path,omitempty"`
	Published *bool   `json:"published,omitempty" yaml:"published,omitempty"`
	Content   *string `json:"-" yaml:"-"`
}

// sortKey is used for sorting.
func (s FServiceVersion) sortKey() string {
	if s.Version != nil {
		return *s.Version
	}
	if s.ID != nil {
		return *s.ID
	}
	return ""
}

// sortKey is used for sorting.
func (s FServicePackage) sortKey() string {
	if s.Name != nil {
		return *s.Name
	}
	if s.ID != nil {
		return *s.ID
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
