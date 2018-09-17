package kong

import "bytes"

// Target represents a Target in Kong.
type Target struct {
	CreatedAt *int64  `json:"created_at,omitempty"`
	ID        *string `json:"id,omitempty"`
	Target    *string `json:"target,omitempty"`
	// TODO change once Upstream/Targets are migrated to new DAO
	UpstreamID *string `json:"upstream_id,omitempty"`
	Weight     *int    `json:"weight,omitempty"`
}

// Valid checks if all the fields in Target are valid.
func (t *Target) Valid() bool {
	if isEmptyString(t.Target) {
		return false
	}
	return true
}

func (t *Target) String() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteByte(' ')
	if isEmptyString(t.ID) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*t.ID)
	}
	buf.WriteByte(' ')
	if isEmptyString(t.Target) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*t.Target)
	}
	buf.WriteByte(' ')
	if isEmptyString(t.UpstreamID) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*t.UpstreamID)
	}
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}
