package kong

import (
	"bytes"
	"strconv"
)

// CIDRPort represents a set of CIDR and a port.
// +k8s:deepcopy-gen=true
type CIDRPort struct {
	IP   *string `json:"ip,omitempty" yaml:"ip,omitempty"`
	Port *int    `json:"port,omitempty" yaml:"port,omitempty"`
}

func (c *CIDRPort) String() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteByte(' ')
	if isEmptyString(c.IP) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.IP)
	}
	buf.WriteByte(' ')
	if c.IP == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(strconv.Itoa(*c.Port))
	}
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}

// Route represents a Route in Kong.
// Read https://getkong.org/docs/0.13.x/admin-api/#Route-object
// +k8s:deepcopy-gen=true
type Route struct {
	CreatedAt     *int        `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	Hosts         []*string   `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	ID            *string     `json:"id,omitempty" yaml:"id,omitempty"`
	Name          *string     `json:"name,omitempty" yaml:"name,omitempty"`
	Methods       []*string   `json:"methods,omitempty" yaml:"methods,omitempty"`
	Paths         []*string   `json:"paths,omitempty" yaml:"paths,omitempty"`
	PreserveHost  *bool       `json:"preserve_host,omitempty" yaml:"preserve_host,omitempty"`
	Protocols     []*string   `json:"protocols,omitempty" yaml:"protocols,omitempty"`
	RegexPriority *int        `json:"regex_priority,omitempty" yaml:"regex_priority,omitempty"`
	Service       *Service    `json:"service,omitempty" yaml:"service,omitempty"`
	StripPath     *bool       `json:"strip_path,omitempty" yaml:"strip_path,omitempty"`
	UpdatedAt     *int        `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	SNIs          []*string   `json:"snis,omitempty" yaml:"snis,omitempty"`
	Sources       []*CIDRPort `json:"sources,omitempty" yaml:"sources,omitempty"`
	Destinations  []*CIDRPort `json:"destinations,omitempty" yaml:"destinations,omitempty"`
	Tags          []*string   `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Valid checks if all the fields in Route are valid.
func (r *Route) Valid() bool {
	if len(r.Protocols) == 0 {
		r.Protocols = StringSlice("http", "https")
	}
	if contains(r.Protocols, "http") || contains(r.Protocols, "https") {
		if len(r.Methods) == 0 && len(r.Paths) == 0 && len(r.Hosts) == 0 {
			return false
		}
	}
	if contains(r.Protocols, "tcp") || contains(r.Protocols, "tls") {
		if len(r.Sources) == 0 && len(r.Destinations) == 0 && len(r.SNIs) == 0 {
			return false
		}
	}
	return true
}

func (r *Route) String() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteByte(' ')
	if isEmptyString(r.ID) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*r.ID)
	}
	buf.WriteByte(' ')
	if isEmptyString(r.Name) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*r.Name)
	}
	buf.WriteByte(' ')
	buf.WriteString(stringArrayToString(r.Methods))
	buf.WriteByte(' ')
	buf.WriteString(stringArrayToString(r.Hosts))
	buf.WriteByte(' ')
	buf.WriteString(stringArrayToString(r.Paths))
	buf.WriteByte(' ')
	buf.WriteString(stringArrayToString(r.SNIs))
	buf.WriteByte(' ')
	buf.WriteString(cidrPortArrayToString(r.Sources))
	buf.WriteByte(' ')
	buf.WriteString(cidrPortArrayToString(r.Destinations))
	buf.WriteByte(' ')
	if r.PreserveHost == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(strconv.FormatBool(*r.PreserveHost))
	}
	buf.WriteByte(' ')
	if r.StripPath == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(strconv.FormatBool(*r.StripPath))
	}
	buf.WriteByte(' ')
	if r.RegexPriority == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(strconv.Itoa(*r.RegexPriority))
	}
	buf.WriteByte(' ')
	if r.Service == nil || isEmptyString(r.Service.ID) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*r.Service.ID)
	}
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}

func cidrPortArrayToString(arr []*CIDRPort) string {
	if arr == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("[ ")
	l := len(arr)
	for i, el := range arr {
		buf.WriteString(el.String())
		if i != l-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString(" ]")
	return buf.String()
}

func contains(slice []*string, s string) bool {
	for _, el := range slice {
		if *el == s {
			return true
		}
	}
	return false
}
