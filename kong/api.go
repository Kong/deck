package kong

// API represents an API in Kong
// Read https://getkong.org/docs/latest/admin-api/#api-object
type API struct {
	CreatedAt              *int64    `json:"created_at,omitempty" yaml:"created_at,omitempty"` //TODO marshal to time.Time
	Hosts                  []*string `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Methods                []*string `json:"methods,omitempty" yaml:"methods,omitempty"` //TODO move to a stricter data type
	URIs                   []*string `json:"uris,omitempty" yaml:"uris,omitempty"`
	HTTPIfTerminated       *bool     `json:"http_if_terminated,omitempty" yaml:"http_if_terminated,omitempty"`
	HTTPSOnly              *bool     `json:"https_only,omitempty" yaml:"https_only,omitempty"`
	ID                     *string   `json:"id,omitempty" yaml:"id,omitempty"`
	Name                   *string   `json:"name,omitempty" yaml:"name,omitempty"`
	PreserveHost           *bool     `json:"preserve_host,omitempty" yaml:"preserve_host,omitempty"`
	Retries                *int      `json:"retries,omitempty" yaml:"retries,omitempty"`
	StripURI               *bool     `json:"strip_uri,omitempty" yaml:"strip_uri,omitempty"`
	UpstreamConnectTimeout *int      `json:"upstream_connect_timeout,omitempty" yaml:"upstream_connect_timeout,omitempty"`
	UpstreamReadTimeout    *int      `json:"upstream_read_timeout,omitempty" yaml:"upstream_read_timeout,omitempty"`
	UpstreamSendTimeout    *int      `json:"upstream_send_timeout,omitempty" yaml:"upstream_send_timeout,omitempty"`
	UpstreamURL            *string   `json:"upstream_url,omitempty" yaml:"upstream_url,omitempty"`
}

// Valid checks if all the fields in API are valid
func (api *API) Valid() bool {
	if isEmptyString(api.Name) || isEmptyString(api.UpstreamURL) {
		return false
	}
	if len(api.Hosts) == 0 && len(api.Methods) == 0 && len(api.URIs) == 0 {
		return false
	}
	// TODO
	// TODO name must only contain alphanumeric and '., -, _, ~' characters
	// TODO check upstreamurl by parsing
	// TODO check methods are valid http methods
	// TODO check URIs starts with /
	// TODO all timeouts must be an integer between 1 and 2147483647
	// TODO "retries": "must be an integer between 0 and 32767"
	// TODO strip all of them
	return true
}
