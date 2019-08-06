package kong

import "bytes"

// Healthy configures thresholds and HTTP status codes
// to mark targets healthy for an upstream.
// +k8s:deepcopy-gen=true
type Healthy struct {
	HTTPStatuses []int `json:"http_statuses,omitempty" yaml:"http_statuses,omitempty"`
	Interval     *int  `json:"interval,omitempty" yaml:"interval,omitempty"`
	Successes    *int  `json:"successes,omitempty" yaml:"successes,omitempty"`
}

// Unhealthy configures thresholds and HTTP status codes
// to mark targets unhealthy.
// +k8s:deepcopy-gen=true
type Unhealthy struct {
	HTTPFailures *int  `json:"http_failures,omitempty" yaml:"http_failures,omitempty"`
	HTTPStatuses []int `json:"http_statuses,omitempty" yaml:"http_statuses,omitempty"`
	TCPFailures  *int  `json:"tcp_failures,omitempty" yaml:"tcp_failures,omitempty"`
	Timeouts     *int  `json:"timeouts,omitempty" yaml:"timeouts,omitempty"`
	Interval     *int  `json:"interval,omitempty" yaml:"interval,omitempty"`
}

// ActiveHealthcheck configures active health check probing.
// +k8s:deepcopy-gen=true
type ActiveHealthcheck struct {
	Concurrency            *int       `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	Healthy                *Healthy   `json:"healthy,omitempty" yaml:"healthy,omitempty"`
	HTTPPath               *string    `json:"http_path,omitempty" yaml:"http_path,omitempty"`
	HTTPSSni               *string    `json:"https_sni,omitempty" yaml:"https_sni,omitempty"`
	HTTPSVerifyCertificate *bool      `json:"https_verify_certificate,omitempty" yaml:"https_verify_certificate,omitempty"`
	Type                   *string    `json:"type,omitempty" yaml:"type,omitempty"`
	Timeout                *int       `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Unhealthy              *Unhealthy `json:"unhealthy,omitempty" yaml:"unhealthy,omitempty"`
}

// PassiveHealthcheck configures passive checks around
// passive health checks.
// +k8s:deepcopy-gen=true
type PassiveHealthcheck struct {
	Healthy   *Healthy   `json:"healthy,omitempty" yaml:"healthy,omitempty"`
	Unhealthy *Unhealthy `json:"unhealthy,omitempty" yaml:"unhealthy,omitempty"`
}

// Healthcheck represents a health-check config of an upstream
// in Kong.
// +k8s:deepcopy-gen=true
type Healthcheck struct {
	Active  *ActiveHealthcheck  `json:"active,omitempty" yaml:"active,omitempty"`
	Passive *PassiveHealthcheck `json:"passive,omitempty" yaml:"passive,omitempty"`
}

// Upstream represents a Consumer in Kong.
// +k8s:deepcopy-gen=true
type Upstream struct {
	ID                 *string      `json:"id,omitempty" yaml:"id,omitempty"`
	Name               *string      `json:"name,omitempty" yaml:"name,omitempty"`
	Slots              *int         `json:"slots,omitempty" yaml:"slots,omitempty"`
	Healthchecks       *Healthcheck `json:"healthchecks,omitempty" yaml:"healthchecks,omitempty"`
	CreatedAt          *int64       `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	HashOn             *string      `json:"hash_on,omitempty" yaml:"hash_on,omitempty"`
	HashFallback       *string      `json:"hash_fallback,omitempty" yaml:"hash_fallback,omitempty"`
	HashOnHeader       *string      `json:"hash_on_header,omitempty" yaml:"hash_on_header,omitempty"`
	HashFallbackHeader *string      `json:"hash_fallback_header,omitempty" yaml:"hash_fallback_header,omitempty"`
	HashOnCookie       *string      `json:"hash_on_cookie,omitempty" yaml:"hash_on_cookie,omitempty"`
	HashOnCookiePath   *string      `json:"hash_on_cookie_path,omitempty" yaml:"hash_on_cookie_path,omitempty"`
	Tags               []*string    `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Valid checks if all the fields in Upstream are valid.
func (u *Upstream) Valid() bool {
	// TODO
	return !isEmptyString(u.Name)
}

func (u *Upstream) String() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteByte(' ')
	if isEmptyString(u.ID) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*u.ID)
	}
	buf.WriteByte(' ')
	if isEmptyString(u.Name) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*u.Name)
	}
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}
