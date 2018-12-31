package state

import (
	"reflect"

	"github.com/hbagdi/go-kong/kong"
)

// Meta contains additional information for an entity
// type Meta struct {
// 	Name   *string `json:"name,omitempty" yaml:"name,omitempty"`
// 	Global *bool   `json:"global,omitempty" yaml:"global,omitempty"`
// 	Kind   *string `json:"type,omitempty" yaml:"type,omitempty"`
// }

// Meta stores metadata for any entity.
type Meta struct {
	metaMap map[string]interface{}
}

func (m *Meta) initMeta() {
	if m.metaMap == nil {
		m.metaMap = make(map[string]interface{})
	}
}

// AddMeta adds key->obj metadata.
// It will override the old obj in key is already present.
func (m *Meta) AddMeta(key string, obj interface{}) {
	m.initMeta()
	m.metaMap[key] = obj
}

// GetMeta returns the obj previously added using AddMeta().
// It returns nil if key is not present.
func (m *Meta) GetMeta(key string) interface{} {
	m.initMeta()
	return m.metaMap[key]
}

// Service represents a service in Kong.
// It adds some helper methods along with Meta to the original Service object.
type Service struct {
	kong.Service `yaml:",inline"`
	Meta
}

// Equal returns true if s1 and s2 are equal.
func (s1 *Service) Equal(s2 *Service) bool {
	return reflect.DeepEqual(s1.Service, s2.Service)
}

// EqualWithOpts returns true if s1 and s2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (s1 *Service) EqualWithOpts(s2 *Service,
	ignoreID bool, ignoreTS bool) bool {
	s1Copy := s1.Service.DeepCopy()
	s2Copy := s2.Service.DeepCopy()

	if ignoreID {
		s1Copy.ID = nil
		s2Copy.ID = nil
	}
	if ignoreTS {
		s1Copy.CreatedAt = nil
		s2Copy.CreatedAt = nil

		s1Copy.UpdatedAt = nil
		s2Copy.UpdatedAt = nil
	}
	return reflect.DeepEqual(s1Copy, s2Copy)
}

// Route represents a route in Kong.
// It adds some helper methods along with Meta to the original Route object.
type Route struct {
	kong.Route `yaml:",inline"`
	Meta
}

// Equal returns true if r1 and r2 are equal.
// TODO add compare array without position
func (r1 *Route) Equal(r2 *Route) bool {
	return reflect.DeepEqual(r1.Route, r2.Route)
}

// EqualWithOpts returns true if r1 and r2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (r1 *Route) EqualWithOpts(r2 *Route, ignoreID,
	ignoreTS, ignoreForeign bool) bool {
	r1Copy := r1.Route.DeepCopy()
	r2Copy := r2.Route.DeepCopy()

	if ignoreID {
		r1Copy.ID = nil
		r2Copy.ID = nil
	}
	if ignoreTS {
		r1Copy.CreatedAt = nil
		r2Copy.CreatedAt = nil

		r1Copy.UpdatedAt = nil
		r2Copy.UpdatedAt = nil
	}
	if ignoreForeign {
		r1Copy.Service = nil
		r2Copy.Service = nil
	}
	return reflect.DeepEqual(r1Copy, r2Copy)
}

// Upstream represents a upstream in Kong.
// It adds some helper methods along with Meta to the original Upstream object.
type Upstream struct {
	kong.Upstream `yaml:",inline"`
	Meta
}

// Equal returns true if u1 and u2 are equal.
func (u1 *Upstream) Equal(u2 *Upstream) bool {
	return reflect.DeepEqual(u1.Upstream, u2.Upstream)
}

// EqualWithOpts returns true if u1 and u2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (u1 *Upstream) EqualWithOpts(u2 *Upstream,
	ignoreID bool, ignoreTS bool) bool {
	u1Copy := u1.Upstream.DeepCopy()
	u2Copy := u2.Upstream.DeepCopy()

	if ignoreID {
		u1Copy.ID = nil
		u2Copy.ID = nil
	}
	if ignoreTS {
		u1Copy.CreatedAt = nil
		u2Copy.CreatedAt = nil
	}
	return reflect.DeepEqual(u1Copy, u2Copy)
}

// Target represents a Target in Kong.
// It adds some helper methods along with Meta to the original Target object.
type Target struct {
	kong.Target `yaml:",inline"`
	Meta
}

// Equal returns true if t1 and t2 are equal.
// TODO add compare array without position
func (t1 *Target) Equal(t2 *Target) bool {
	return reflect.DeepEqual(t1.Target, t2.Target)
}

// EqualWithOpts returns true if t1 and t2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (t1 *Target) EqualWithOpts(t2 *Target, ignoreID,
	ignoreTS, ignoreForeign bool) bool {
	t1Copy := t1.Target.DeepCopy()
	t2Copy := t2.Target.DeepCopy()

	if ignoreID {
		t1Copy.ID = nil
		t2Copy.ID = nil
	}
	if ignoreTS {
		t1Copy.CreatedAt = nil
		t2Copy.CreatedAt = nil
	}
	if ignoreForeign {
		t1Copy.Upstream = nil
		t2Copy.Upstream = nil
	}
	return reflect.DeepEqual(t1Copy, t2Copy)
}
