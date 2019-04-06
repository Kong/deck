package kong

import "bytes"

// Consumer represents a Consumer in Kong.
// Read https://getkong.org/docs/0.13.x/admin-api/#consumer-object
// +k8s:deepcopy-gen=true
type Consumer struct {
	ID        *string   `json:"id,omitempty" yaml:"id,omitempty"`
	CustomID  *string   `json:"custom_id,omitempty" yaml:"custom_id,omitempty"`
	Username  *string   `json:"username,omitempty" yaml:"username,omitempty"`
	CreatedAt *int64    `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	Tags      []*string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Valid checks if all the fields in Consumer are valid.
func (c *Consumer) Valid() bool {
	emptyCustomID := isEmptyString(c.CustomID)
	emptyUsername := isEmptyString(c.Username)

	return !(emptyCustomID && emptyUsername)

}

func (c *Consumer) String() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteByte(' ')
	if isEmptyString(c.ID) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.ID)
	}
	buf.WriteByte(' ')
	if isEmptyString(c.Username) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.Username)
	}
	buf.WriteByte(' ')
	if isEmptyString(c.CustomID) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.CustomID)
	}
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}
