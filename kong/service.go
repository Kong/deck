package kong

import (
	"bytes"
	"strconv"
)

// Service represents a Service in Kong.
// Read https://getkong.org/docs/0.13.x/admin-api/#Service-object
// +k8s:deepcopy-gen=true
type Service struct {
	ClientCertificate *Certificate `json:"client_certificate,omitempty" yamls:"client_certificate"`
	ConnectTimeout    *int         `json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"`
	CreatedAt         *int         `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	Host              *string      `json:"host,omitempty" yaml:"host,omitempty"`
	ID                *string      `json:"id,omitempty" yaml:"id,omitempty"`
	Name              *string      `json:"name,omitempty" yaml:"name,omitempty"`
	Path              *string      `json:"path,omitempty" yaml:"path,omitempty"`
	Port              *int         `json:"port,omitempty" yaml:"port,omitempty"`
	Protocol          *string      `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	ReadTimeout       *int         `json:"read_timeout,omitempty" yaml:"read_timeout,omitempty"`
	Retries           *int         `json:"retries,omitempty" yaml:"retries,omitempty"`
	UpdatedAt         *int         `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	WriteTimeout      *int         `json:"write_timeout,omitempty" yaml:"write_timeout,omitempty"`
	Tags              []*string    `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func (s *Service) String() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteByte(' ')
	if isEmptyString(s.ID) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*s.ID)
	}
	buf.WriteByte(' ')
	if isEmptyString(s.Name) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*s.Name)
	}
	buf.WriteByte(' ')
	if isEmptyString(s.Protocol) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*s.Protocol)
	}
	buf.WriteByte(' ')
	if isEmptyString(s.Host) {
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
	if isEmptyString(s.Path) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*s.Path)
	}
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}
