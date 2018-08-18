package kong

import (
	"bytes"
	"strconv"
	"strings"
)

// Service represents a Service in Kong.
// Read https://getkong.org/docs/0.13.x/admin-api/#Service-object
type Service struct {
	ConnectTimeout *int    `json:"connect_timeout"`
	CreatedAt      *int    `json:"created_at"`
	Host           *string `json:"host"`
	ID             *string `json:"id"`
	Name           *string `json:"name"`
	Path           *string `json:"path"`
	Port           *int    `json:"port"`
	Protocol       *string `json:"protocol"`
	ReadTimeout    *int    `json:"read_timeout"`
	Retries        *int    `json:"retries"`
	UpdatedAt      *int    `json:"updated_at"`
	WriteTimeout   *int    `json:"write_timeout"`
}

// Valid checks if all the fields in Service are valid.
func (s *Service) Valid() bool {
	if s.Protocol != nil &&
		strings.ToLower(*s.Protocol) != "http" &&
		strings.ToLower(*s.Protocol) != "https" {
		return false
	}
	return true
}

func (s *Service) String() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteByte(' ')
	if s.ID == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*s.ID)
	}
	buf.WriteByte(' ')
	if s.Name == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*s.Name)
	}
	buf.WriteByte(' ')
	if s.Protocol == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*s.Protocol)
	}
	buf.WriteByte(' ')
	if s.Host == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*s.Host)
	}
	buf.WriteByte(' ')
	if s.Port == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(strconv.Itoa(*s.Port))
	}
	buf.WriteByte(' ')
	if s.Path == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*s.Path)
	}
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}
