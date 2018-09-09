package kong

import "bytes"

// Healthy configures thresholds and HTTP status codes
// to mark targets healthy for an upstream.
type Healthy struct {
	HTTPStatuses []int `json:"http_statuses,omitempty"`
	Interval     *int  `json:"interval,omitempty"`
	Successes    *int  `json:"successes,omitempty"`
}

// Unhealthy configures thresholds and HTTP status codes
// to mark targets unhealthy.
type Unhealthy struct {
	HTTPFailures *int  `json:"http_failures,omitempty"`
	HTTPStatuses []int `json:"http_statuses,omitempty"`
	TCPFailures  *int  `json:"tcp_failures,omitempty"`
	Timeouts     *int  `json:"timeouts,omitempty"`
}

// ActiveHealthcheck configures active health check probing.
type ActiveHealthcheck struct {
	Concurrency *int       `json:"concurrency,omitempty"`
	Healthy     *Healthy   `json:"healthy,omitempty"`
	HTTPPath    *string    `json:"http_path,omitempty"`
	Timeout     *int       `json:"timeout,omitempty"`
	Unhealthy   *Unhealthy `json:"unhealthy,omitempty"`
}

// PassiveHealthcheck configures passive checks around
// passive health checks.
type PassiveHealthcheck struct {
	Healthy   *Healthy   `json:"healthy,omitempty"`
	Unhealthy *Unhealthy `json:"unhealthy,omitempty"`
}

// Healthcheck represents a health-check config of an upstream
// in Kong.
type Healthcheck struct {
	Active  *ActiveHealthcheck  `json:"active,omitempty"`
	Passive *PassiveHealthcheck `json:"passive,omitempty"`
}

// Upstream represents a Consumer in Kong.
type Upstream struct {
	CreatedAt        *int64      `json:"created_at,omitempty"`
	HashFallback     *string     `json:"hash_fallback,omitempty"`
	HashOn           *string     `json:"hash_on,omitempty"`
	HashOnCookiePath *string     `json:"hash_on_cookie_path,omitempty"`
	Healthchecks     Healthcheck `json:"healthchecks,omitempty"`
	ID               *string     `json:"id,omitempty"`
	Name             *string     `json:"name,omitempty"`
	Slots            *int        `json:"slots,omitempty"`
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
	if u.ID == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*u.ID)
	}
	buf.WriteByte(' ')
	if u.Name == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*u.Name)
	}
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}
