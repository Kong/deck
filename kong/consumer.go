package kong

import "bytes"

// Consumer represents a Consumer in Kong.
// Read https://getkong.org/docs/0.13.x/admin-api/#consumer-object
type Consumer struct {
	ID        *string `json:"id"`
	CustomID  *string `json:"custom_id,omitempty"`
	Username  *string `json:"username,omitempty"`
	CreatedAt *int64  `json:"created_at"`
}

// Valid checks if all the fields in Consumer are valid.
func (c *Consumer) Valid() bool {
	emptyCustomID := isEmptyString(c.CustomID)
	emptyUsername := isEmptyString(c.Username)

	return emptyCustomID != emptyUsername

}

func (c *Consumer) String() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteByte(' ')
	if c.ID == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.ID)
	}
	buf.WriteByte(' ')
	if c.Username == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.Username)
	}
	buf.WriteByte(' ')
	if c.CustomID == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*c.CustomID)
	}
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}
