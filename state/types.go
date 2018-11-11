package state

import (
	"reflect"

	"github.com/hbagdi/go-kong/kong"
)

// Meta contains additional information for an entity
type Meta struct {
	Name   *string `json:"name,omitempty" yaml:"name,omitempty"`
	Global *bool   `json:"global,omitempty" yaml:"global,omitempty"`
	Kind   *string `json:"type,omitempty" yaml:"type,omitempty"`
}

// Service represents a service in Kong with helper methods
type Service struct {
	kong.Service `yaml:",inline"`
	Meta         map[string]interface{}
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
}

// TODO add compare array without position
func (s *Route) Equal(s2 *Route) bool {
	return reflect.DeepEqual(s, s2)
}

func (s *Route) EqualWithOpts(s2 *Route, ignoreID bool, ignoreTS bool) bool {
	sCopy := s.DeepCopy()
	s2Copy := s2.DeepCopy()

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

// can be used for reading in state
type ServiceNode struct {
	kong.Service
	Meta *Meta
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
