package kong

import (
	"bytes"
	"fmt"
)

// Configuration represents a config of a plugin in Kong.
type Configuration map[string]interface{}

// Plugin represents a Plugin in Kong.
// Read https://getkong.org/docs/0.13.x/admin-api/#Plugin-object
type Plugin struct {
	CreatedAt  *int          `json:"created_at,omitempty"`
	ID         *string       `json:"id,omitempty"`
	Name       *string       `json:"name,omitempty"`
	RouteID    *string       `json:"route_id,omitempty"`
	ServiceID  *string       `json:"service_id,omitempty"`
	APIID      *string       `json:"api_id,omitempty"`
	ConsumerID *string       `json:"consumer_id,omitempty"`
	Config     Configuration `json:"config,omitempty"`
	Enabled    *bool         `json:"enabled,omitempty"`
}

// Valid checks if all the fields in Plugin are valid.
func (p *Plugin) Valid() bool {
	return !isEmptyString(p.Name)
}

func (p *Plugin) String() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteByte(' ')
	if p.ID == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*p.ID)
	}
	buf.WriteByte(' ')
	if p.Name == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*p.Name)
	}
	buf.WriteByte(' ')
	if p.RouteID == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*p.RouteID)
	}
	buf.WriteByte(' ')
	if p.ServiceID == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*p.ServiceID)
	}
	buf.WriteByte(' ')
	if p.APIID == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*p.APIID)
	}
	buf.WriteByte(' ')
	if p.ConsumerID == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*p.ConsumerID)
	}
	buf.WriteByte(' ')
	buf.WriteString(fmt.Sprint(p.Config))
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}
