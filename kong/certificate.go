package kong

import "bytes"

// Certificate represents a Certificate in Kong.
// Read https://getkong.org/docs/0.14.x/admin-api/#certificate-object
// +k8s:deepcopy-gen=true
type Certificate struct {
	ID        *string   `json:"id,omitempty" yaml:"id,omitempty"`
	Cert      *string   `json:"cert,omitempty" yaml:"cert,omitempty"`
	Key       *string   `json:"key,omitempty" yaml:"key,omitempty"`
	CreatedAt *int64    `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	SNIs      []*string `json:"snis,omitempty" yaml:"snis,omitempty"`
	Tags      []*string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Valid checks if all the fields in Consumer are valid.
func (c *Certificate) Valid() bool {
	// TODO
	return true
}

func (c *Certificate) String() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteByte(' ')
	if isEmptyString(c.ID) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.ID)
	}
	buf.WriteByte(' ')
	buf.WriteString(stringArrayToString(c.SNIs))
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}
