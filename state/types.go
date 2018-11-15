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

type Meta struct {
	metaMap map[string]interface{}
}

func (m *Meta) initMeta() {
	if m.metaMap == nil {
		m.metaMap = make(map[string]interface{})
	}
}

func (m *Meta) AddMeta(key string, obj interface{}) {
	m.initMeta()
	m.metaMap["key"] = obj
}

func (m *Meta) GetMeta(key string) interface{} {
	m.initMeta()
	return m.metaMap["key"]
}

// Service represents a service in Kong with helper methods
type Service struct {
	kong.Service `yaml:",inline"`
	Meta
}

func (s *Service) Equal(s2 *Service) bool {
	return reflect.DeepEqual(s.Service, s2.Service)
}

func (s *Service) EqualWithOpts(s2 *Service, ignoreID bool, ignoreTS bool) bool {
	sCopy := s.Service.DeepCopy()
	s2Copy := s2.Service.DeepCopy()

	if ignoreID {
		sCopy.ID = nil
		s2Copy.ID = nil
	}
	if ignoreTS {
		sCopy.CreatedAt = nil
		s2Copy.CreatedAt = nil

		sCopy.UpdatedAt = nil
		s2Copy.UpdatedAt = nil
	}
	return reflect.DeepEqual(sCopy, s2Copy)
}

type Route struct {
	kong.Route `yaml:",inline"`
	Meta       map[string]interface{}
}

// TODO add compare array without position
func (r1 *Route) Equal(r2 *Route) bool {
	return reflect.DeepEqual(r1.Route, r2.Route)
}

func (r *Route) EqualWithOpts(r2 *Route, ignoreID bool, ignoreTS bool) bool {
	rCopy := r.Route.DeepCopy()
	r2Copy := r2.Route.DeepCopy()

	if ignoreID {
		rCopy.ID = nil
		r2Copy.ID = nil
	}
	if ignoreTS {
		rCopy.CreatedAt = nil
		r2Copy.CreatedAt = nil

		rCopy.UpdatedAt = nil
		r2Copy.UpdatedAt = nil
	}
	return reflect.DeepEqual(rCopy, r2Copy)
}

// can be used for reading in state
type ServiceNode struct {
	kong.Service
	Meta
	// Routes  []*Route
	// Plugins []*Plugin
}

// type Route struct {
// 	kong.Route
// }

// type RouteNode struct {
// 	kong.Route
// 	Meta    *Meta
// 	Plugins []*Plugin
// }

// type Plugin struct {
// 	kong.Plugin
// 	Meta *Meta
// }
