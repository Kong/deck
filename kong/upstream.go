package kong

import "bytes"

// Healthy configures thresholds and HTTP status codes
// to mark targets healthy for an upstream.
type Healthy struct {
	HTTPStatuses []int `json:"http_statuses"`
	Interval     *int  `json:"interval"`
	Successes    *int  `json:"successes"`
}

// Unhealthy configures thresholds and HTTP status codes
// to mark targets unhealthy.
type Unhealthy struct {
	HTTPFailures *int  `json:"http_failures"`
	HTTPStatuses []int `json:"http_statuses"`
	TCPFailures  *int  `json:"tcp_failures"`
	Timeouts     *int  `json:"timeouts"`
}

// ActiveHealthcheck configures active health check probing.
type ActiveHealthcheck struct {
	Concurrency *int       `json:"concurrency"`
	Healthy     *Healthy   `json:"healthy"`
	HTTPPath    *string    `json:"http_path"`
	Timeout     *int       `json:"timeout"`
	Unhealthy   *Unhealthy `json:"unhealthy"`
}

// PassiveHealthcheck configures passive checks around
// passive health checks.
type PassiveHealthcheck struct {
	Healthy   *Healthy   `json:"healthy"`
	Unhealthy *Unhealthy `json:"unhealthy"`
}

// Healthcheck represents a health-check config of an upstream
// in Kong.
type Healthcheck struct {
	Active  *ActiveHealthcheck  `json:"active"`
	Passive *PassiveHealthcheck `json:"passive"`
}

// Upstream represents a Consumer in Kong.
type Upstream struct {
	CreatedAt        *int64      `json:"created_at"`
	HashFallback     *string     `json:"hash_fallback"`
	HashOn           *string     `json:"hash_on"`
	HashOnCookiePath *string     `json:"hash_on_cookie_path"`
	Healthchecks     Healthcheck `json:"healthchecks"`
	ID               *string     `json:"id"`
	Name             *string     `json:"name"`
	Slots            *int        `json:"slots"`
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
