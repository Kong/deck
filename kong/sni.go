package kong

import "bytes"

// SNI represents a SNI in Kong.
// Read https://getkong.org/docs/0.14.x/admin-api/#sni-object
// +k8s:deepcopy-gen=true
type SNI struct {
	ID          *string      `json:"id,omitempty" yaml:"id,omitempty"`
	Name        *string      `json:"name,omitempty" yaml:"name,omitempty"`
	CreatedAt   *int64       `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	Certificate *Certificate `json:"certificate,omitempty" yaml:"certificate,omitempty"`
	Tags        []*string    `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Valid checks if all the fields in SNI are valid.
func (c *SNI) Valid() bool {
	if isEmptyString(c.Name) {
		return false
	}
	return true
}

func (c *SNI) String() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteByte(' ')
	if isEmptyString(c.ID) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.ID)
	}
	buf.WriteByte(' ')
	if isEmptyString(c.Name) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.Name)
	}
	buf.WriteByte(' ')
	if c.Certificate == nil || isEmptyString(c.Certificate.ID) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.Certificate.ID)
	}
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}
