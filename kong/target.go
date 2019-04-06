package kong

import "bytes"

// Target represents a Target in Kong.
// +k8s:deepcopy-gen=true
type Target struct {
	CreatedAt *float64  `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	ID        *string   `json:"id,omitempty" yaml:"id,omitempty"`
	Target    *string   `json:"target,omitempty" yaml:"target,omitempty"`
	Upstream  *Upstream `json:"upstream,omitempty" yaml:"upstream,omitempty"`
	Weight    *int      `json:"weight,omitempty" yaml:"weight,omitempty"`
	Tags      []*string `json:"tags,omitempty" yaml:"tags,omitempty"`
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
	if t.Upstream == nil || isEmptyString(t.Upstream.ID) {
		buf.WriteString("nil")
	} else {
		buf.WriteString(*t.Upstream.ID)
	}
	buf.WriteByte(' ')
	buf.WriteByte(']')
	return buf.String()
}
