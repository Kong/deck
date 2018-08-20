package kong

import "bytes"

// Certificate represents a Certificate in Kong.
// Read https://getkong.org/docs/0.14.x/admin-api/#certificate-object
type Certificate struct {
	ID        *string   `json:"id"`
	Cert      *string   `json:"cert,omitempty"`
	Key       *string   `json:"key,omitempty"`
	CreatedAt *int64    `json:"created_at,omitempty"`
	SNIs      []*string `json:"snis,omitempty"`
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
	if c.ID == nil {
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
