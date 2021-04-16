package state

import (
	"github.com/kong/deck/konnect"
	"reflect"
)

// Document represents a document in Konnect.
// It adds some helper methods along with Meta to the original Document object.
type Document struct {
	konnect.Document `yaml:",inline"`
	Meta
}

// Console returns an entity's identity in a human-readable string.
func (d *Document) Console() string {
	return *d.Path
}

// ServicePackage represents a service package in Konnect.
// It adds some helper methods along with Meta to the original ServicePackage object.
type ServicePackage struct {
	konnect.ServicePackage `yaml:",inline"`
	Meta
}

// Identifier returns the endpoint key name or ID.
func (s1 *ServicePackage) Identifier() string {
	if s1.Name != nil {
		return *s1.Name
	}
	return *s1.ID
}

// Console returns an entity's identity in a human
// readable string.
func (s1 *ServicePackage) Console() string {
	return s1.Identifier()
}

// Equal returns true if s1 and s2 are equal.
func (s1 *ServicePackage) Equal(s2 *ServicePackage) bool {
	return s1.EqualWithOpts(s2, false, false)
}

// EqualWithOpts returns true if s1 and s2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (s1 *ServicePackage) EqualWithOpts(s2 *ServicePackage,
	ignoreID bool, ignoreTS bool) bool {
	s1Copy := s1.ServicePackage.DeepCopy()
	s2Copy := s2.ServicePackage.DeepCopy()

	if ignoreID {
		s1Copy.ID = nil
		s2Copy.ID = nil
	}
	return reflect.DeepEqual(s1Copy, s2Copy)
}

// ServiceVersion represents a service version in Konnect.
// It adds some helper methods along with Meta to the original ServiceVersion
//object.
type ServiceVersion struct {
	konnect.ServiceVersion `yaml:",inline"`
	Meta
}

// Identifier returns the endpoint key name or ID.
func (s1 *ServiceVersion) Identifier() string {
	if s1.Version != nil {
		return *s1.Version
	}
	return *s1.ID
}

// Console returns an entity's identity in a human
// readable string.
func (s1 *ServiceVersion) Console() string {
	return s1.Identifier()
}

// Equal returns true if s1 and s2 are equal.
func (s1 *ServiceVersion) Equal(s2 *ServiceVersion) bool {
	return s1.EqualWithOpts(s2, false, false, false)
}

// EqualWithOpts returns true if s1 and s2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (s1 *ServiceVersion) EqualWithOpts(s2 *ServiceVersion,
	ignoreID, ignoreTS, ignoreForeign bool) bool {
	s1Copy := s1.ServiceVersion.DeepCopy()
	s2Copy := s2.ServiceVersion.DeepCopy()

	if ignoreID {
		s1Copy.ID = nil
		s2Copy.ID = nil
	}
	if ignoreForeign {
		s1Copy.ServicePackage = nil
		s1Copy.ControlPlaneServiceRelation = nil
		s2Copy.ServicePackage = nil
		s2Copy.ControlPlaneServiceRelation = nil
	}
	return reflect.DeepEqual(s1Copy, s2Copy)
}
