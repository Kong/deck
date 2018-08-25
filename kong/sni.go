package kong

import "bytes"

// SNI represents a SNI in Kong.
// Read https://getkong.org/docs/0.14.x/admin-api/#sni-object
type SNI struct {
	ID          *string      `json:"id"`
	Name        *string      `json:"name,omitempty"`
	CreatedAt   *int64       `json:"created_at,omitempty"`
	Certificate *Certificate `json:"certificate,omitempty"`
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
	if c.ID == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.ID)
	}
	buf.WriteByte(' ')
	if c.Name == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.Name)
	}
	buf.WriteByte(' ')
	if c.Certificate == nil || c.Certificate.ID == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.Certificate.ID)
	}
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}
