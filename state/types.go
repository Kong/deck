package state

import (
	"reflect"

	"github.com/hbagdi/go-kong/kong"
)

// entity abstracts out common fields in a credentials.
// TODO generalize for each and every entity.
type entity interface {
	// ID of the cred.
	GetID() string
	// ID2 is the second endpoint key.
	GetID2() string
	// Consumer returns consumer ID associated with the cred.
	GetConsumer() string
}

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

// Certificate represents a upstream in Kong.
// It adds some helper methods along with Meta to the
// original Certificate object.
type Certificate struct {
	kong.Certificate `yaml:",inline"`
	Meta
}

// Equal returns true if c1 and c2 are equal.
func (c1 *Certificate) Equal(c2 *Certificate) bool {
	return reflect.DeepEqual(c1.Certificate, c2.Certificate)
}

// EqualWithOpts returns true if c1 and c2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (c1 *Certificate) EqualWithOpts(c2 *Certificate,
	ignoreID bool, ignoreTS bool) bool {
	c1Copy := c1.Certificate.DeepCopy()
	c2Copy := c2.Certificate.DeepCopy()

	if ignoreID {
		c1Copy.ID = nil
		c2Copy.ID = nil
	}
	if ignoreTS {
		c1Copy.CreatedAt = nil
		c2Copy.CreatedAt = nil
	}
	return reflect.DeepEqual(c1Copy, c2Copy)
}

// SNI represents a SNI in Kong.
// It adds some helper methods along with Meta to the original SNI object.
type SNI struct {
	kong.SNI `yaml:",inline"`
	Meta
}

// Equal returns true if s1 and s2 are equal.
// TODO add compare array without position
func (s1 *SNI) Equal(s2 *SNI) bool {
	return reflect.DeepEqual(s1.SNI, s2.SNI)
}

// EqualWithOpts returns true if s1 and s2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (s1 *SNI) EqualWithOpts(s2 *SNI, ignoreID,
	ignoreTS, ignoreForeign bool) bool {
	s1Copy := s1.SNI.DeepCopy()
	s2Copy := s2.SNI.DeepCopy()

	if ignoreID {
		s1Copy.ID = nil
		s2Copy.ID = nil
	}
	if ignoreTS {
		s1Copy.CreatedAt = nil
		s2Copy.CreatedAt = nil
	}
	if ignoreForeign {
		s1Copy.Certificate = nil
		s2Copy.Certificate = nil
	}
	return reflect.DeepEqual(s1Copy, s2Copy)
}

// Plugin represents a route in Kong.
// It adds some helper methods along with Meta to the original Plugin object.
type Plugin struct {
	kong.Plugin `yaml:",inline"`
	Meta
}

// Equal returns true if r1 and r2 are equal.
// TODO add compare array without position
func (p1 *Plugin) Equal(p2 *Plugin) bool {
	return reflect.DeepEqual(p1.Plugin, p2.Plugin)
}

// EqualWithOpts returns true if p1 and p2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (p1 *Plugin) EqualWithOpts(p2 *Plugin, ignoreID,
	ignoreTS, ignoreForeign bool) bool {
	p1Copy := p1.Plugin.DeepCopy()
	p2Copy := p2.Plugin.DeepCopy()

	if ignoreID {
		p1Copy.ID = nil
		p2Copy.ID = nil
	}
	if ignoreTS {
		p1Copy.CreatedAt = nil
		p2Copy.CreatedAt = nil
	}
	if ignoreForeign {
		p1Copy.Service = nil
		p1Copy.Route = nil
		p1Copy.Consumer = nil
		p2Copy.Service = nil
		p2Copy.Route = nil
		p2Copy.Consumer = nil
	}
	return reflect.DeepEqual(p1Copy, p2Copy)
}

// Consumer represents a consumer in Kong.
// It adds some helper methods along with Meta to the original Consumer object.
type Consumer struct {
	kong.Consumer `yaml:",inline"`
	Meta
}

// Equal returns true if c1 and c2 are equal.
func (c1 *Consumer) Equal(c2 *Consumer) bool {
	return reflect.DeepEqual(c1.Consumer, c2.Consumer)
}

// EqualWithOpts returns true if c1 and c2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (c1 *Consumer) EqualWithOpts(c2 *Consumer,
	ignoreID bool, ignoreTS bool) bool {
	c1Copt := c1.Consumer.DeepCopy()
	c2Copy := c2.Consumer.DeepCopy()

	if ignoreID {
		c1Copt.ID = nil
		c2Copy.ID = nil
	}
	if ignoreTS {
		c1Copt.CreatedAt = nil
		c2Copy.CreatedAt = nil
	}
	return reflect.DeepEqual(c1Copt, c2Copy)
}

// KeyAuth represents a key-auth credential in Kong.
// It adds some helper methods along with Meta to the original KeyAuth object.
type KeyAuth struct {
	kong.KeyAuth `yaml:",inline"`
	Meta
}

// Equal returns true if k1 and k2 are equal.
func (k1 *KeyAuth) Equal(k2 *KeyAuth) bool {
	return reflect.DeepEqual(k1.KeyAuth, k2.KeyAuth)
}

// EqualWithOpts returns true if k1 and k2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (k1 *KeyAuth) EqualWithOpts(k2 *KeyAuth, ignoreID,
	ignoreTS, ignoreForeign bool) bool {
	k1Copy := k1.KeyAuth.DeepCopy()
	k2Copy := k2.KeyAuth.DeepCopy()

	if ignoreID {
		k1Copy.ID = nil
		k2Copy.ID = nil
	}
	if ignoreTS {
		k1Copy.CreatedAt = nil
		k2Copy.CreatedAt = nil
	}
	if ignoreForeign {
		k1Copy.Consumer = nil
		k2Copy.Consumer = nil
	}
	return reflect.DeepEqual(k1Copy, k2Copy)
}

// GetID returns ID.
// If ID is empty, it returns an empty string.
func (k1 *KeyAuth) GetID() string {
	if k1.ID == nil {
		return ""
	}
	return *k1.ID
}

// GetID2 returns the endpoint key of the entity,
// the Key field for KeyAuth.
func (k1 *KeyAuth) GetID2() string {
	if k1.Key == nil {
		return ""
	}
	return *k1.Key
}

// GetConsumer returns the credential's Consumer's ID.
// If Consumer's ID is empty, it returns an empty string.
func (k1 *KeyAuth) GetConsumer() string {
	if k1.Consumer == nil || k1.Consumer.ID == nil {
		return ""
	}
	return *k1.Consumer.ID
}

// HMACAuth represents a key-auth credential in Kong.
// It adds some helper methods along with Meta to the original HMACAuth object.
type HMACAuth struct {
	kong.HMACAuth `yaml:",inline"`
	Meta
}

// Equal returns true if h1 and h2 are equal.
func (h1 *HMACAuth) Equal(h2 *HMACAuth) bool {
	return reflect.DeepEqual(h1.HMACAuth, h2.HMACAuth)
}

// EqualWithOpts returns true if h1 and h2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (h1 *HMACAuth) EqualWithOpts(h2 *HMACAuth, ignoreID,
	ignoreTS, ignoreForeign bool) bool {
	h1Copy := h1.HMACAuth.DeepCopy()
	h2Copy := h2.HMACAuth.DeepCopy()

	if ignoreID {
		h1Copy.ID = nil
		h2Copy.ID = nil
	}
	if ignoreTS {
		h1Copy.CreatedAt = nil
		h2Copy.CreatedAt = nil
	}
	if ignoreForeign {
		h1Copy.Consumer = nil
		h2Copy.Consumer = nil
	}
	return reflect.DeepEqual(h1Copy, h2Copy)
}

// GetID returns ID.
// If ID is empty, it returns an empty string.
func (h1 *HMACAuth) GetID() string {
	if h1.ID == nil {
		return ""
	}
	return *h1.ID
}

// GetID2 returns the endpoint key of the entity,
// the Username field for HMACAuth.
func (h1 *HMACAuth) GetID2() string {
	if h1.Username == nil {
		return ""
	}
	return *h1.Username
}

// GetConsumer returns the credential's Consumer's ID.
// If Consumer's ID is empty, it returns an empty string.
func (h1 *HMACAuth) GetConsumer() string {
	if h1.Consumer == nil || h1.Consumer.ID == nil {
		return ""
	}
	return *h1.Consumer.ID
}

// JWTAuth represents a jwt credential in Kong.
// It adds some helper methods along with Meta to the original JWTAuth object.
type JWTAuth struct {
	kong.JWTAuth `yaml:",inline"`
	Meta
}

// Equal returns true if j1 and j2 are equal.
func (j1 *JWTAuth) Equal(j2 *JWTAuth) bool {
	return reflect.DeepEqual(j1.JWTAuth, j2.JWTAuth)
}

// EqualWithOpts returns true if j1 and j2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (j1 *JWTAuth) EqualWithOpts(j2 *JWTAuth, ignoreID,
	ignoreTS, ignoreForeign bool) bool {
	j1Copy := j1.JWTAuth.DeepCopy()
	j2Copy := j2.JWTAuth.DeepCopy()

	if ignoreID {
		j1Copy.ID = nil
		j2Copy.ID = nil
	}
	if ignoreTS {
		j1Copy.CreatedAt = nil
		j2Copy.CreatedAt = nil
	}
	if ignoreForeign {
		j1Copy.Consumer = nil
		j2Copy.Consumer = nil
	}
	return reflect.DeepEqual(j1Copy, j2Copy)
}

// BasicAuth represents a basic-auth credential in Kong.
// It adds some helper methods along with Meta to the original BasicAuth object.
type BasicAuth struct {
	kong.BasicAuth `yaml:",inline"`
	Meta
}

// Equal returns true if b1 and b2 are equal.
func (b1 *BasicAuth) Equal(b2 *BasicAuth) bool {
	return reflect.DeepEqual(b1.BasicAuth, b2.BasicAuth)
}

// EqualWithOpts returns true if j1 and j2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (b1 *BasicAuth) EqualWithOpts(b2 *BasicAuth, ignoreID,
	ignoreTS, ignorePassword, ignoreForeign bool) bool {
	b1Copy := b1.BasicAuth.DeepCopy()
	b2Copy := b2.BasicAuth.DeepCopy()

	if ignoreID {
		b1Copy.ID = nil
		b2Copy.ID = nil
	}
	if ignoreTS {
		b1Copy.CreatedAt = nil
		b2Copy.CreatedAt = nil
	}
	if ignorePassword {
		b1Copy.Password = nil
		b2Copy.Password = nil
	}
	if ignoreForeign {
		b1Copy.Consumer = nil
		b2Copy.Consumer = nil
	}
	return reflect.DeepEqual(b1Copy, b2Copy)
}

// GetID returns ID.
// If ID is empty, it returns an empty string.
func (k1 *BasicAuth) GetID() string {
	if k1.ID == nil {
		return ""
	}
	return *k1.ID
}

// GetID2 returns the endpoint key of the entity,
// the Username field for BasicAuth.
func (k1 *BasicAuth) GetID2() string {
	if k1.Username == nil {
		return ""
	}
	return *k1.Username
}

// GetConsumer returns the credential's Consumer's ID.
// If Consumer's ID is empty, it returns an empty string.
func (k1 *BasicAuth) GetConsumer() string {
	if k1.Consumer == nil || k1.Consumer.ID == nil {
		return ""
	}
	return *k1.Consumer.ID
}

// ACLGroup represents an ACL group for a consumer in Kong.
// It adds some helper methods along with Meta to the original ACLGroup object.
type ACLGroup struct {
	kong.ACLGroup `yaml:",inline"`
	Meta
}

// Equal returns true if b1 and b2 are equal.
func (b1 *ACLGroup) Equal(b2 *ACLGroup) bool {
	return reflect.DeepEqual(b1.ACLGroup, b2.ACLGroup)
}

// EqualWithOpts returns true if j1 and j2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (b1 *ACLGroup) EqualWithOpts(b2 *ACLGroup, ignoreID,
	ignoreTS, ignoreForeign bool) bool {
	b1Copy := b1.ACLGroup.DeepCopy()
	b2Copy := b2.ACLGroup.DeepCopy()

	if ignoreID {
		b1Copy.ID = nil
		b2Copy.ID = nil
	}
	if ignoreTS {
		b1Copy.CreatedAt = nil
		b2Copy.CreatedAt = nil
	}
	if ignoreForeign {
		b1Copy.Consumer = nil
		b2Copy.Consumer = nil
	}
	return reflect.DeepEqual(b1Copy, b2Copy)
}

// CACertificate represents a CACertificate in Kong.
// It adds some helper methods along with Meta to the
// original CACertificate object.
type CACertificate struct {
	kong.CACertificate `yaml:",inline"`
	Meta
}

// Equal returns true if c1 and c2 are equal.
func (c1 *CACertificate) Equal(c2 *CACertificate) bool {
	return reflect.DeepEqual(c1.CACertificate, c2.CACertificate)
}

// EqualWithOpts returns true if c1 and c2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (c1 *CACertificate) EqualWithOpts(c2 *CACertificate,
	ignoreID bool, ignoreTS bool) bool {
	c1Copy := c1.CACertificate.DeepCopy()
	c2Copy := c2.CACertificate.DeepCopy()

	if ignoreID {
		c1Copy.ID = nil
		c2Copy.ID = nil
	}
	if ignoreTS {
		c1Copy.CreatedAt = nil
		c2Copy.CreatedAt = nil
	}
	return reflect.DeepEqual(c1Copy, c2Copy)
}

// Oauth2Credential represents an Oauth2 credential in Kong.
// It adds some helper methods along with Meta to the original Oauth2Credential object.
type Oauth2Credential struct {
	kong.Oauth2Credential `yaml:",inline"`
	Meta
}

// Equal returns true if k1 and k2 are equal.
func (k1 *Oauth2Credential) Equal(k2 *Oauth2Credential) bool {
	return reflect.DeepEqual(k1.Oauth2Credential, k2.Oauth2Credential)
}

// EqualWithOpts returns true if k1 and k2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (k1 *Oauth2Credential) EqualWithOpts(k2 *Oauth2Credential, ignoreID,
	ignoreTS, ignoreForeign bool) bool {
	k1Copy := k1.Oauth2Credential.DeepCopy()
	k2Copy := k2.Oauth2Credential.DeepCopy()

	if ignoreID {
		k1Copy.ID = nil
		k2Copy.ID = nil
	}
	if ignoreTS {
		k1Copy.CreatedAt = nil
		k2Copy.CreatedAt = nil
	}
	if ignoreForeign {
		k1Copy.Consumer = nil
		k2Copy.Consumer = nil
	}
	return reflect.DeepEqual(k1Copy, k2Copy)
}
