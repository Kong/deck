package state

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/kong/go-kong/kong"
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

// ConsoleString contains methods to be used to print
// entity to console.
type ConsoleString interface {
	// Console returns a string to uniquely identify an
	// entity in human-readable form.
	// It should have the ID or endpoint key along-with
	// foreign references if they exist.
	// It will be used to communicate to the human user
	// that this entity is undergoing some change.
	Console() string
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

// Identifier returns the endpoint key name or ID.
func (s1 *Service) Identifier() string {
	if s1.Name != nil {
		return *s1.Name
	}
	return *s1.ID
}

// Console returns an entity's identity in a human
// readable string.
func (s1 *Service) Console() string {
	return s1.FriendlyName()
}

// Equal returns true if s1 and s2 are equal.
func (s1 *Service) Equal(s2 *Service) bool {
	return s1.EqualWithOpts(s2, false, false)
}

// EqualWithOpts returns true if s1 and s2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (s1 *Service) EqualWithOpts(s2 *Service,
	ignoreID bool, ignoreTS bool,
) bool {
	s1Copy := s1.Service.DeepCopy()
	s2Copy := s2.Service.DeepCopy()

	if len(s1Copy.Tags) == 0 {
		s1Copy.Tags = nil
	}
	if len(s2Copy.Tags) == 0 {
		s2Copy.Tags = nil
	}

	// Cassandra can sometimes mess up tag order, but tag order doesn't actually matter: tags are sets
	// even though we represent them with slices. Sort before comparison to avoid spurious diff detection.
	sort.Slice(s1Copy.Tags, func(i, j int) bool { return *(s1Copy.Tags[i]) < *(s1Copy.Tags[j]) })
	sort.Slice(s2Copy.Tags, func(i, j int) bool { return *(s2Copy.Tags[i]) < *(s2Copy.Tags[j]) })

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

// Identifier returns the endpoint key name or ID.
func (r1 *Route) Identifier() string {
	if r1.Name != nil {
		return *r1.Name
	}
	return *r1.ID
}

// Console returns an entity's identity in a human
// readable string.
func (r1 *Route) Console() string {
	return r1.FriendlyName()
}

// Equal returns true if r1 and r2 are equal.
// TODO add compare array without position
func (r1 *Route) Equal(r2 *Route) bool {
	return r1.EqualWithOpts(r2, false, false, false)
}

// EqualWithOpts returns true if r1 and r2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (r1 *Route) EqualWithOpts(r2 *Route, ignoreID,
	ignoreTS, ignoreForeign bool,
) bool {
	r1Copy := r1.Route.DeepCopy()
	r2Copy := r2.Route.DeepCopy()

	if len(r1Copy.Tags) == 0 {
		r1Copy.Tags = nil
	}
	if len(r2Copy.Tags) == 0 {
		r2Copy.Tags = nil
	}

	sort.Slice(r1Copy.Tags, func(i, j int) bool { return *(r1Copy.Tags[i]) < *(r1Copy.Tags[j]) })
	sort.Slice(r2Copy.Tags, func(i, j int) bool { return *(r2Copy.Tags[i]) < *(r2Copy.Tags[j]) })

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

	if r1Copy.Service != nil {
		r1Copy.Service.Name = nil
	}
	if r2Copy.Service != nil {
		r2Copy.Service.Name = nil
	}
	return reflect.DeepEqual(r1Copy, r2Copy)
}

// Upstream represents a upstream in Kong.
// It adds some helper methods along with Meta to the original Upstream object.
type Upstream struct {
	kong.Upstream `yaml:",inline"`
	Meta
}

// Identifier returns the endpoint key name or ID.
func (u1 *Upstream) Identifier() string {
	if u1.Name != nil {
		return *u1.Name
	}
	return *u1.ID
}

// Console returns an entity's identity in a human
// readable string.
func (u1 *Upstream) Console() string {
	return u1.FriendlyName()
}

// Equal returns true if u1 and u2 are equal.
func (u1 *Upstream) Equal(u2 *Upstream) bool {
	return u1.EqualWithOpts(u2, false, false)
}

// EqualWithOpts returns true if u1 and u2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (u1 *Upstream) EqualWithOpts(u2 *Upstream,
	ignoreID bool, ignoreTS bool,
) bool {
	u1Copy := u1.Upstream.DeepCopy()
	u2Copy := u2.Upstream.DeepCopy()

	if len(u1Copy.Tags) == 0 {
		u1Copy.Tags = nil
	}
	if len(u2Copy.Tags) == 0 {
		u2Copy.Tags = nil
	}

	sort.Slice(u1Copy.Tags, func(i, j int) bool { return *(u1Copy.Tags[i]) < *(u1Copy.Tags[j]) })
	sort.Slice(u2Copy.Tags, func(i, j int) bool { return *(u2Copy.Tags[i]) < *(u2Copy.Tags[j]) })

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

// Identifier returns the endpoint key name or ID.
func (t1 *Target) Identifier() string {
	if t1.Target.Target != nil {
		return *t1.Target.Target
	}
	return *t1.ID
}

// Console returns an entity's identity in a human
// readable string.
func (t1 *Target) Console() string {
	res := t1.FriendlyName()
	if t1.Upstream != nil {
		res = res + " for upstream " + t1.Upstream.FriendlyName()
	}
	return res
}

// Equal returns true if t1 and t2 are equal.
// TODO add compare array without position
func (t1 *Target) Equal(t2 *Target) bool {
	return t1.EqualWithOpts(t2, false, false, false)
}

// EqualWithOpts returns true if t1 and t2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (t1 *Target) EqualWithOpts(t2 *Target, ignoreID,
	ignoreTS, ignoreForeign bool,
) bool {
	t1Copy := t1.Target.DeepCopy()
	t2Copy := t2.Target.DeepCopy()

	if len(t1Copy.Tags) == 0 {
		t1Copy.Tags = nil
	}
	if len(t2Copy.Tags) == 0 {
		t2Copy.Tags = nil
	}

	sort.Slice(t1Copy.Tags, func(i, j int) bool { return *(t1Copy.Tags[i]) < *(t1Copy.Tags[j]) })
	sort.Slice(t2Copy.Tags, func(i, j int) bool { return *(t2Copy.Tags[i]) < *(t2Copy.Tags[j]) })

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

// Identifier returns the endpoint key name or ID.
func (c1 *Certificate) Identifier() string {
	if c1.ID != nil {
		return *c1.ID
	}
	return *c1.Cert
}

// Console returns an entity's identity in a human
// readable string.
func (c1 *Certificate) Console() string {
	return c1.FriendlyName()
}

// Equal returns true if c1 and c2 are equal.
func (c1 *Certificate) Equal(c2 *Certificate) bool {
	return c1.EqualWithOpts(c2, false, false)
}

// EqualWithOpts returns true if c1 and c2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (c1 *Certificate) EqualWithOpts(c2 *Certificate,
	ignoreID bool, ignoreTS bool,
) bool {
	c1Copy := c1.Certificate.DeepCopy()
	c2Copy := c2.Certificate.DeepCopy()

	if len(c1Copy.Tags) == 0 {
		c1Copy.Tags = nil
	}
	if len(c2Copy.Tags) == 0 {
		c2Copy.Tags = nil
	}

	sort.Slice(c1Copy.Tags, func(i, j int) bool { return *(c1Copy.Tags[i]) < *(c1Copy.Tags[j]) })
	sort.Slice(c2Copy.Tags, func(i, j int) bool { return *(c2Copy.Tags[i]) < *(c2Copy.Tags[j]) })

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

// Identifier returns the endpoint key name or ID.
func (s1 *SNI) Identifier() string {
	if s1.Name != nil {
		return *s1.Name
	}
	return *s1.ID
}

// Equal returns true if s1 and s2 are equal.
// TODO add compare array without position
func (s1 *SNI) Equal(s2 *SNI) bool {
	return s1.EqualWithOpts(s2, false, false, false)
}

// Console returns an entity's identity in a human
// readable string.
func (s1 *SNI) Console() string {
	return s1.FriendlyName()
}

// EqualWithOpts returns true if s1 and s2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (s1 *SNI) EqualWithOpts(s2 *SNI, ignoreID,
	ignoreTS, ignoreForeign bool,
) bool {
	s1Copy := s1.SNI.DeepCopy()
	s2Copy := s2.SNI.DeepCopy()

	if len(s1Copy.Tags) == 0 {
		s1Copy.Tags = nil
	}
	if len(s2Copy.Tags) == 0 {
		s2Copy.Tags = nil
	}

	sort.Slice(s1Copy.Tags, func(i, j int) bool { return *(s1Copy.Tags[i]) < *(s1Copy.Tags[j]) })
	sort.Slice(s2Copy.Tags, func(i, j int) bool { return *(s2Copy.Tags[i]) < *(s2Copy.Tags[j]) })

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

// Identifier returns the endpoint key name or ID.
func (p1 *Plugin) Identifier() string {
	if p1.Name != nil {
		return *p1.Name
	}
	return *p1.ID
}

// Console returns an entity's identity in a human
// readable string.
func (p1 *Plugin) Console() string {
	res := *p1.Name + " "

	if p1.Service == nil && p1.Route == nil && p1.Consumer == nil {
		return res + "(global)"
	}
	associations := []string{}
	if p1.Service != nil {
		associations = append(associations, "service "+p1.Service.FriendlyName())
	}
	if p1.Route != nil {
		associations = append(associations, "route "+p1.Route.FriendlyName())
	}
	if p1.Consumer != nil {
		associations = append(associations, "consumer "+p1.Consumer.FriendlyName())
	}
	if p1.ConsumerGroup != nil {
		associations = append(associations, "consumer-group "+p1.ConsumerGroup.FriendlyName())
	}
	if len(associations) > 0 {
		res += "for "
	}
	for i := 0; i < len(associations); i++ {
		res += associations[i]
		if i < len(associations)-1 {
			res += " and "
		}
	}
	return res
}

// Equal returns true if r1 and r2 are equal.
// TODO add compare array without position
func (p1 *Plugin) Equal(p2 *Plugin) bool {
	return p1.EqualWithOpts(p2, false, false, false)
}

// EqualWithOpts returns true if p1 and p2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (p1 *Plugin) EqualWithOpts(p2 *Plugin, ignoreID,
	ignoreTS, ignoreForeign bool,
) bool {
	p1Copy := p1.Plugin.DeepCopy()
	p2Copy := p2.Plugin.DeepCopy()

	if len(p1Copy.Tags) == 0 {
		p1Copy.Tags = nil
	}
	if len(p2Copy.Tags) == 0 {
		p2Copy.Tags = nil
	}

	sort.Slice(p1Copy.Tags, func(i, j int) bool { return *(p1Copy.Tags[i]) < *(p1Copy.Tags[j]) })
	sort.Slice(p2Copy.Tags, func(i, j int) bool { return *(p2Copy.Tags[i]) < *(p2Copy.Tags[j]) })

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
		p2Copy.ConsumerGroup = nil
	}

	if p1Copy.Service != nil {
		p1Copy.Service.Name = nil
	}
	if p2Copy.Service != nil {
		p2Copy.Service.Name = nil
	}
	if p1Copy.Route != nil {
		p1Copy.Route.Name = nil
	}
	if p2Copy.Route != nil {
		p2Copy.Route.Name = nil
	}
	if p1Copy.Consumer != nil {
		p1Copy.Consumer.Username = nil
	}
	if p2Copy.Consumer != nil {
		p2Copy.Consumer.Username = nil
	}
	if p1Copy.ConsumerGroup != nil {
		p1Copy.ConsumerGroup.Name = nil
	}
	if p2Copy.ConsumerGroup != nil {
		p2Copy.ConsumerGroup.Name = nil
	}
	return reflect.DeepEqual(p1Copy, p2Copy)
}

// Consumer represents a consumer in Kong.
// It adds some helper methods along with Meta to the original Consumer object.
type Consumer struct {
	kong.Consumer `yaml:",inline"`
	Meta
}

// Identifier returns the endpoint key name or ID.
func (c1 *Consumer) Identifier() string {
	if c1.Username != nil {
		return *c1.Username
	}
	return *c1.ID
}

// Console returns an entity's identity in a human
// readable string.
func (c1 *Consumer) Console() string {
	return c1.FriendlyName()
}

// Equal returns true if c1 and c2 are equal.
func (c1 *Consumer) Equal(c2 *Consumer) bool {
	return c1.EqualWithOpts(c2, false, false)
}

// EqualWithOpts returns true if c1 and c2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (c1 *Consumer) EqualWithOpts(c2 *Consumer,
	ignoreID bool, ignoreTS bool,
) bool {
	c1Copy := c1.Consumer.DeepCopy()
	c2Copy := c2.Consumer.DeepCopy()

	if len(c1Copy.Tags) == 0 {
		c1Copy.Tags = nil
	}
	if len(c2Copy.Tags) == 0 {
		c2Copy.Tags = nil
	}

	sort.Slice(c1Copy.Tags, func(i, j int) bool { return *(c1Copy.Tags[i]) < *(c1Copy.Tags[j]) })
	sort.Slice(c2Copy.Tags, func(i, j int) bool { return *(c2Copy.Tags[i]) < *(c2Copy.Tags[j]) })

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

func forConsumerString(c *kong.Consumer) string {
	if c != nil {
		friendlyName := c.FriendlyName()
		if friendlyName != "" {
			return " for consumer " + friendlyName
		}
	}
	return ""
}

// ConsumerGroupObject represents a ConsumerGroupObject in Kong.
// It adds some helper methods along with Meta to the original Upstream object.
type ConsumerGroupObject struct {
	kong.ConsumerGroupObject `yaml:",inline"`
	Meta
}

// Identifier returns the endpoint key name or ID.
func (c1 *ConsumerGroupObject) Identifier() string {
	if c1.ConsumerGroup != nil && c1.ConsumerGroup.Name != nil {
		return *c1.ConsumerGroup.Name
	}
	return *c1.ConsumerGroup.ID
}

// Console returns an entity's identity in a human
// readable string.
func (c1 *ConsumerGroupObject) Console() string {
	return c1.ConsumerGroup.FriendlyName()
}

// Equal returns true if u1 and u2 are equal.
func (c1 *ConsumerGroupObject) Equal(c2 *ConsumerGroupObject) bool {
	return c1.EqualWithOpts(c2, false, false)
}

// EqualWithOpts returns true if c1 and c2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (c1 *ConsumerGroupObject) EqualWithOpts(c2 *ConsumerGroupObject,
	ignoreID bool, ignoreTS bool,
) bool {
	c1Copy := c1.ConsumerGroup.DeepCopy()
	c2Copy := c2.ConsumerGroup.DeepCopy()

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

// ConsumerGroup represents a ConsumerGroup in Kong.
// It adds some helper methods along with Meta to the original ConsumerGroup object.
type ConsumerGroup struct {
	kong.ConsumerGroup `yaml:",inline"`
	Meta
}

// Identifier returns the endpoint key name or ID.
func (c1 *ConsumerGroup) Identifier() string {
	if c1.ConsumerGroup.Name != nil {
		return *c1.ConsumerGroup.Name
	}
	return *c1.ConsumerGroup.ID
}

// Console returns an entity's identity in a human
// readable string.
func (c1 *ConsumerGroup) Console() string {
	return c1.ConsumerGroup.FriendlyName()
}

// Equal returns true if c1 and c2 are equal.
func (c1 *ConsumerGroup) Equal(c2 *ConsumerGroup) bool {
	return c1.EqualWithOpts(c2, false, false)
}

// EqualWithOpts returns true if c1 and c2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (c1 *ConsumerGroup) EqualWithOpts(c2 *ConsumerGroup,
	ignoreID bool, ignoreTS bool,
) bool {
	u1Copy := c1.ConsumerGroup.DeepCopy()
	u2Copy := c2.ConsumerGroup.DeepCopy()

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

// ConsumerGroupConsumer represents a ConsumerGroupConsumer in Kong.
// It adds some helper methods along with Meta to the original ConsumerGroupConsumer object.
type ConsumerGroupConsumer struct {
	kong.ConsumerGroupConsumer `yaml:",inline"`
	Meta
}

// Identifier returns the endpoint key Ursername or ID.
func (c1 *ConsumerGroupConsumer) Identifier() string {
	if c1.Consumer.Username != nil {
		return *c1.Consumer.Username
	}
	return *c1.Consumer.ID
}

// Console returns an entity's identity in a human
// readable string.
func (c1 *ConsumerGroupConsumer) Console() string {
	return *c1.ConsumerGroupConsumer.Consumer.Username
}

// Equal returns true if c1 and c2 are equal.
func (c1 *ConsumerGroupConsumer) Equal(c2 *ConsumerGroupConsumer) bool {
	return c1.EqualWithOpts(c2, false, false)
}

// EqualWithOpts returns true if c1 and c2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (c1 *ConsumerGroupConsumer) EqualWithOpts(c2 *ConsumerGroupConsumer,
	ignoreID bool, ignoreTS bool,
) bool {
	c1Copy := c1.ConsumerGroupConsumer.DeepCopy()
	c2Copy := c2.ConsumerGroupConsumer.DeepCopy()
	if ignoreID {
		c1Copy.Consumer.ID = nil
		c2Copy.Consumer.ID = nil
	}
	if ignoreTS {
		c1Copy.CreatedAt = nil
		c2Copy.CreatedAt = nil
		c1Copy.Consumer.CreatedAt = nil
		c2Copy.Consumer.CreatedAt = nil
		c2Copy.ConsumerGroup.CreatedAt = nil
		c1Copy.ConsumerGroup.CreatedAt = nil
	}
	return reflect.DeepEqual(c1Copy, c2Copy)
}

// ConsumerGroupPlugin represents a ConsumerGroupConsumer in Kong.
// It adds some helper methods along with Meta to the original ConsumerGroupConsumer object.
type ConsumerGroupPlugin struct {
	kong.ConsumerGroupPlugin `yaml:",inline"`
	Meta
}

// Identifier returns the endpoint key name or ID.
func (c1 *ConsumerGroupPlugin) Identifier() string {
	if c1.Name != nil {
		return *c1.Name
	}
	return *c1.ID
}

// Console returns an entity's identity in a human
// readable string.
func (c1 *ConsumerGroupPlugin) Console() string {
	return *c1.Name
}

// Equal returns true if c1 and c2 are equal.
func (c1 *ConsumerGroupPlugin) Equal(c2 *ConsumerGroupPlugin) bool {
	return c1.EqualWithOpts(c2, false, false)
}

// EqualWithOpts returns true if c1 and c2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (c1 *ConsumerGroupPlugin) EqualWithOpts(c2 *ConsumerGroupPlugin,
	ignoreID bool, ignoreTS bool,
) bool {
	c1Copy := c1.DeepCopy()
	c2Copy := c2.DeepCopy()
	if ignoreID {
		c1Copy.ID = nil
		c2Copy.ID = nil
	}
	if ignoreTS {
		c1Copy.CreatedAt = nil
		c2Copy.CreatedAt = nil
		c1Copy.ConsumerGroup.CreatedAt = nil
		c2Copy.ConsumerGroup.CreatedAt = nil
	}
	return reflect.DeepEqual(c1Copy, c2Copy)
}

// KeyAuth represents a key-auth credential in Kong.
// It adds some helper methods along with Meta to the original KeyAuth object.
type KeyAuth struct {
	kong.KeyAuth `yaml:",inline"`
	Meta
}

// stripKey returns the last 5 characters of key.
// If key is less than or equal to 5 characters, then the key is returned as is.
func stripKey(key string) string {
	const keyIdentifierLength = 5
	if len(key) <= keyIdentifierLength {
		return key
	}
	return key[len(key)-keyIdentifierLength:]
}

// Console returns an entity's identity in a human
// readable string.
func (k1 *KeyAuth) Console() string {
	return stripKey(*k1.Key) + forConsumerString(k1.Consumer)
}

// Equal returns true if k1 and k2 are equal.
func (k1 *KeyAuth) Equal(k2 *KeyAuth) bool {
	return k1.EqualWithOpts(k2, false, false, false)
}

// EqualWithOpts returns true if k1 and k2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (k1 *KeyAuth) EqualWithOpts(k2 *KeyAuth, ignoreID,
	ignoreTS, ignoreForeign bool,
) bool {
	k1Copy := k1.KeyAuth.DeepCopy()
	k2Copy := k2.KeyAuth.DeepCopy()

	if len(k1Copy.Tags) == 0 {
		k1Copy.Tags = nil
	}
	if len(k2Copy.Tags) == 0 {
		k2Copy.Tags = nil
	}

	sort.Slice(k1Copy.Tags, func(i, j int) bool { return *(k1Copy.Tags[i]) < *(k1Copy.Tags[j]) })
	sort.Slice(k2Copy.Tags, func(i, j int) bool { return *(k2Copy.Tags[i]) < *(k2Copy.Tags[j]) })

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
	if k1Copy.Consumer != nil {
		k1Copy.Consumer.Username = nil
	}
	if k2Copy.Consumer != nil {
		k2Copy.Consumer.Username = nil
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

// Console returns an entity's identity in a human
// readable string.
func (h1 *HMACAuth) Console() string {
	return *h1.Username + forConsumerString(h1.Consumer)
}

// Equal returns true if h1 and h2 are equal.
func (h1 *HMACAuth) Equal(h2 *HMACAuth) bool {
	return h1.EqualWithOpts(h2, false, false, false)
}

// EqualWithOpts returns true if h1 and h2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (h1 *HMACAuth) EqualWithOpts(h2 *HMACAuth, ignoreID,
	ignoreTS, ignoreForeign bool,
) bool {
	h1Copy := h1.HMACAuth.DeepCopy()
	h2Copy := h2.HMACAuth.DeepCopy()

	if len(h1Copy.Tags) == 0 {
		h1Copy.Tags = nil
	}
	if len(h2Copy.Tags) == 0 {
		h2Copy.Tags = nil
	}

	sort.Slice(h1Copy.Tags, func(i, j int) bool { return *(h1Copy.Tags[i]) < *(h1Copy.Tags[j]) })
	sort.Slice(h2Copy.Tags, func(i, j int) bool { return *(h2Copy.Tags[i]) < *(h2Copy.Tags[j]) })

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
	if h1Copy.Consumer != nil {
		h1Copy.Consumer.Username = nil
	}
	if h2Copy.Consumer != nil {
		h2Copy.Consumer.Username = nil
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

// Console returns an entity's identity in a human
// readable string.
func (j1 *JWTAuth) Console() string {
	return *j1.Key + forConsumerString(j1.Consumer)
}

// Equal returns true if j1 and j2 are equal.
func (j1 *JWTAuth) Equal(j2 *JWTAuth) bool {
	return j1.EqualWithOpts(j2, false, false, false)
}

// EqualWithOpts returns true if j1 and j2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (j1 *JWTAuth) EqualWithOpts(j2 *JWTAuth, ignoreID,
	ignoreTS, ignoreForeign bool,
) bool {
	j1Copy := j1.JWTAuth.DeepCopy()
	j2Copy := j2.JWTAuth.DeepCopy()

	if len(j1Copy.Tags) == 0 {
		j1Copy.Tags = nil
	}
	if len(j2Copy.Tags) == 0 {
		j2Copy.Tags = nil
	}

	sort.Slice(j1Copy.Tags, func(i, j int) bool { return *(j1Copy.Tags[i]) < *(j1Copy.Tags[j]) })
	sort.Slice(j2Copy.Tags, func(i, j int) bool { return *(j2Copy.Tags[i]) < *(j2Copy.Tags[j]) })

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
	if j1Copy.Consumer != nil {
		j1Copy.Consumer.Username = nil
	}
	if j2Copy.Consumer != nil {
		j2Copy.Consumer.Username = nil
	}
	return reflect.DeepEqual(j1Copy, j2Copy)
}

// GetID returns ID.
// If ID is empty, it returns an empty string.
func (j1 *JWTAuth) GetID() string {
	if j1.ID == nil {
		return ""
	}
	return *j1.ID
}

// GetID2 returns the endpoint key of the entity,
// the Key field for JWTAuth.
func (j1 *JWTAuth) GetID2() string {
	if j1.Key == nil {
		return ""
	}
	return *j1.Key
}

// GetConsumer returns the credential's Consumer's ID.
// If Consumer's ID is empty, it returns an empty string.
func (j1 *JWTAuth) GetConsumer() string {
	if j1.Consumer == nil || j1.Consumer.ID == nil {
		return ""
	}
	return *j1.Consumer.ID
}

// BasicAuth represents a basic-auth credential in Kong.
// It adds some helper methods along with Meta to the original BasicAuth object.
type BasicAuth struct {
	kong.BasicAuth `yaml:",inline"`
	Meta
}

// Console returns an entity's identity in a human
// readable string.
func (b1 *BasicAuth) Console() string {
	return *b1.Username + forConsumerString(b1.Consumer)
}

// Equal returns true if b1 and b2 are equal.
func (b1 *BasicAuth) Equal(b2 *BasicAuth) bool {
	return b1.EqualWithOpts(b2, false, false, false, false)
}

// EqualWithOpts returns true if j1 and j2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (b1 *BasicAuth) EqualWithOpts(b2 *BasicAuth, ignoreID,
	ignoreTS, ignorePassword, ignoreForeign bool,
) bool {
	b1Copy := b1.BasicAuth.DeepCopy()
	b2Copy := b2.BasicAuth.DeepCopy()

	if len(b1Copy.Tags) == 0 {
		b1Copy.Tags = nil
	}
	if len(b2Copy.Tags) == 0 {
		b2Copy.Tags = nil
	}

	sort.Slice(b1Copy.Tags, func(i, j int) bool { return *(b1Copy.Tags[i]) < *(b1Copy.Tags[j]) })
	sort.Slice(b2Copy.Tags, func(i, j int) bool { return *(b2Copy.Tags[i]) < *(b2Copy.Tags[j]) })

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
	if b1Copy.Consumer != nil {
		b1Copy.Consumer.Username = nil
	}
	if b2Copy.Consumer != nil {
		b2Copy.Consumer.Username = nil
	}
	return reflect.DeepEqual(b1Copy, b2Copy)
}

// GetID returns ID.
// If ID is empty, it returns an empty string.
func (b1 *BasicAuth) GetID() string {
	if b1.ID == nil {
		return ""
	}
	return *b1.ID
}

// GetID2 returns the endpoint key of the entity,
// the Username field for BasicAuth.
func (b1 *BasicAuth) GetID2() string {
	if b1.Username == nil {
		return ""
	}
	return *b1.Username
}

// GetConsumer returns the credential's Consumer's ID.
// If Consumer's ID is empty, it returns an empty string.
func (b1 *BasicAuth) GetConsumer() string {
	if b1.Consumer == nil || b1.Consumer.ID == nil {
		return ""
	}
	return *b1.Consumer.ID
}

// ACLGroup represents an ACL group for a consumer in Kong.
// It adds some helper methods along with Meta to the original ACLGroup object.
type ACLGroup struct {
	kong.ACLGroup `yaml:",inline"`
	Meta
}

// Console returns an entity's identity in a human
// readable string.
func (b1 *ACLGroup) Console() string {
	return *b1.Group + forConsumerString(b1.Consumer)
}

// Equal returns true if b1 and b2 are equal.
func (b1 *ACLGroup) Equal(b2 *ACLGroup) bool {
	return b1.EqualWithOpts(b2, false, false, false)
}

// EqualWithOpts returns true if j1 and j2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (b1 *ACLGroup) EqualWithOpts(b2 *ACLGroup, ignoreID,
	ignoreTS, ignoreForeign bool,
) bool {
	b1Copy := b1.ACLGroup.DeepCopy()
	b2Copy := b2.ACLGroup.DeepCopy()

	if len(b1Copy.Tags) == 0 {
		b1Copy.Tags = nil
	}
	if len(b2Copy.Tags) == 0 {
		b2Copy.Tags = nil
	}

	sort.Slice(b1Copy.Tags, func(i, j int) bool { return *(b1Copy.Tags[i]) < *(b1Copy.Tags[j]) })
	sort.Slice(b2Copy.Tags, func(i, j int) bool { return *(b2Copy.Tags[i]) < *(b2Copy.Tags[j]) })

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
	if b1Copy.Consumer != nil {
		b1Copy.Consumer.Username = nil
	}
	if b2Copy.Consumer != nil {
		b2Copy.Consumer.Username = nil
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

// Identifier returns the endpoint key name or ID.
func (c1 *CACertificate) Identifier() string {
	if c1.ID != nil {
		return *c1.ID
	}
	return *c1.Cert
}

// Console returns an entity's identity in a human
// readable string.
func (c1 *CACertificate) Console() string {
	return c1.FriendlyName()
}

// Equal returns true if c1 and c2 are equal.
func (c1 *CACertificate) Equal(c2 *CACertificate) bool {
	return c1.EqualWithOpts(c2, false, false)
}

// EqualWithOpts returns true if c1 and c2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (c1 *CACertificate) EqualWithOpts(c2 *CACertificate,
	ignoreID bool, ignoreTS bool,
) bool {
	c1Copy := c1.CACertificate.DeepCopy()
	c2Copy := c2.CACertificate.DeepCopy()

	if len(c1Copy.Tags) == 0 {
		c1Copy.Tags = nil
	}
	if len(c2Copy.Tags) == 0 {
		c2Copy.Tags = nil
	}

	sort.Slice(c1Copy.Tags, func(i, j int) bool { return *(c1Copy.Tags[i]) < *(c1Copy.Tags[j]) })
	sort.Slice(c2Copy.Tags, func(i, j int) bool { return *(c2Copy.Tags[i]) < *(c2Copy.Tags[j]) })

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

// Console returns an entity's identity in a human
// readable string.
func (k1 *Oauth2Credential) Console() string {
	return *k1.Name + forConsumerString(k1.Consumer)
}

// Equal returns true if k1 and k2 are equal.
func (k1 *Oauth2Credential) Equal(k2 *Oauth2Credential) bool {
	return k1.EqualWithOpts(k2, false, false, false)
}

// EqualWithOpts returns true if k1 and k2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (k1 *Oauth2Credential) EqualWithOpts(k2 *Oauth2Credential, ignoreID,
	ignoreTS, ignoreForeign bool,
) bool {
	k1Copy := k1.Oauth2Credential.DeepCopy()
	k2Copy := k2.Oauth2Credential.DeepCopy()

	if len(k1Copy.Tags) == 0 {
		k1Copy.Tags = nil
	}
	if len(k2Copy.Tags) == 0 {
		k2Copy.Tags = nil
	}

	sort.Slice(k1Copy.Tags, func(i, j int) bool { return *(k1Copy.Tags[i]) < *(k1Copy.Tags[j]) })
	sort.Slice(k2Copy.Tags, func(i, j int) bool { return *(k2Copy.Tags[i]) < *(k2Copy.Tags[j]) })

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
	if k1Copy.Consumer != nil {
		k1Copy.Consumer.Username = nil
	}
	if k2Copy.Consumer != nil {
		k2Copy.Consumer.Username = nil
	}
	return reflect.DeepEqual(k1Copy, k2Copy)
}

// GetID returns ID.
// If ID is empty, it returns an empty string.
func (k1 *Oauth2Credential) GetID() string {
	if k1.ID == nil {
		return ""
	}
	return *k1.ID
}

// GetID2 returns the endpoint key of the entity,
// the ClientID field for Oauth2Credential.
func (k1 *Oauth2Credential) GetID2() string {
	if k1.ClientID == nil {
		return ""
	}
	return *k1.ClientID
}

// GetConsumer returns the credential's Consumer's ID.
// If Consumer's ID is empty, it returns an empty string.
func (k1 *Oauth2Credential) GetConsumer() string {
	if k1.Consumer == nil || k1.Consumer.ID == nil {
		return ""
	}
	return *k1.Consumer.ID
}

// MTLSAuth represents an mtls-auth credential in Kong.
// It adds some helper methods along with Meta to the original MTLSAuth object.
type MTLSAuth struct {
	kong.MTLSAuth `yaml:",inline"`
	Meta
}

// Console returns an entity's identity in a human
// readable string.
func (b1 *MTLSAuth) Console() string {
	return *b1.SubjectName + forConsumerString(b1.Consumer)
}

// Equal returns true if b1 and b2 are equal.
func (b1 *MTLSAuth) Equal(b2 *MTLSAuth) bool {
	return b1.EqualWithOpts(b2, false, false, false)
}

// EqualWithOpts returns true if j1 and j2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (b1 *MTLSAuth) EqualWithOpts(b2 *MTLSAuth, ignoreID,
	ignoreTS, ignoreForeign bool,
) bool {
	b1Copy := b1.MTLSAuth.DeepCopy()
	b2Copy := b2.MTLSAuth.DeepCopy()

	if len(b1Copy.Tags) == 0 {
		b1Copy.Tags = nil
	}
	if len(b2Copy.Tags) == 0 {
		b2Copy.Tags = nil
	}

	sort.Slice(b1Copy.Tags, func(i, j int) bool { return *(b1Copy.Tags[i]) < *(b1Copy.Tags[j]) })
	sort.Slice(b2Copy.Tags, func(i, j int) bool { return *(b2Copy.Tags[i]) < *(b2Copy.Tags[j]) })

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
	if b1Copy.Consumer != nil {
		b1Copy.Consumer.Username = nil
	}
	if b2Copy.Consumer != nil {
		b2Copy.Consumer.Username = nil
	}
	return reflect.DeepEqual(b1Copy, b2Copy)
}

// RBACRole represents an RBAC Role in Kong.
// It adds some helper methods along with Meta to the original RBACRole object.
type RBACRole struct {
	kong.RBACRole `yaml:",inline"`
	Meta
}

// Identifier returns the endpoint key name or ID.
func (r1 *RBACRole) Identifier() string {
	if r1.Name != nil {
		return *r1.Name
	}
	return *r1.ID
}

// Console returns an entity's identity in a human
// readable string.
func (r1 *RBACRole) Console() string {
	return r1.FriendlyName()
}

// Equal returns true if r1 and r2 are equal.
// TODO add compare array without position
func (r1 *RBACRole) Equal(r2 *RBACRole) bool {
	return r1.EqualWithOpts(r2, false, false, false)
}

// EqualWithOpts returns true if r1 and r2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (r1 *RBACRole) EqualWithOpts(r2 *RBACRole, ignoreID,
	ignoreTS, _ bool,
) bool {
	r1Copy := r1.RBACRole.DeepCopy()
	r2Copy := r2.RBACRole.DeepCopy()

	if ignoreID {
		r1Copy.ID = nil
		r2Copy.ID = nil
	}
	if ignoreTS {
		r1Copy.CreatedAt = nil
		r2Copy.CreatedAt = nil
	}

	return reflect.DeepEqual(r1Copy, r2Copy)
}

// RBACEndpointPermission represents an RBAC Role in Kong.
// It adds some helper methods along with Meta to the original RBACEndpointPermission object.
type RBACEndpointPermission struct {
	ID                          string
	kong.RBACEndpointPermission `yaml:",inline"`
	Meta
}

// Identifier returns a composite ID base on Role ID, workspace, and endpoint
func (r1 *RBACEndpointPermission) Identifier() string {
	if r1.Endpoint != nil {
		return fmt.Sprintf("%s-%s-%s", *r1.Role.ID, *r1.Workspace, *r1.Endpoint)
	}
	return *r1.Endpoint
}

// Console returns an entity's identity in a human
// readable string.
func (r1 *RBACEndpointPermission) Console() string {
	return r1.FriendlyName()
}

// Equal returns true if r1 and r2 are equal.
// TODO add compare array without position
func (r1 *RBACEndpointPermission) Equal(r2 *RBACEndpointPermission) bool {
	return r1.EqualWithOpts(r2, false, false, false)
}

// EqualWithOpts returns true if r1 and r2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (r1 *RBACEndpointPermission) EqualWithOpts(r2 *RBACEndpointPermission, ignoreID,
	ignoreTS, _ bool,
) bool {
	r1Copy := r1.RBACEndpointPermission.DeepCopy()
	r2Copy := r2.RBACEndpointPermission.DeepCopy()

	if ignoreID {
		r1Copy.Endpoint = nil
		r2Copy.Endpoint = nil
	}
	if ignoreTS {
		r1Copy.CreatedAt = nil
		r2Copy.CreatedAt = nil
	}

	return reflect.DeepEqual(r1Copy, r2Copy)
}

// GetID returns ID.
// If ID is empty, it returns an empty string.
func (b1 *MTLSAuth) GetID() string {
	if b1.ID == nil {
		return ""
	}
	return *b1.ID
}

// GetID2 returns the endpoint key of the entity,
// BUT NO SUCH THING EXISTS 😱
// TODO: this is kind of a pointless clone of GetID for MTLSAuth. the mtls-auth
// entity cannot be referenced by anything other than its ID (it has no unique
// fields), but the entity interface requires this function. this duplication
// doesn't appear to be harmful, but it's weird.
func (b1 *MTLSAuth) GetID2() string {
	return (*b1).GetID()
}

func (b1 *MTLSAuth) GetConsumer() string {
	if b1.Consumer == nil || b1.Consumer.ID == nil {
		return ""
	}
	return *b1.Consumer.ID
}

// Vault represents a vault in Kong.
// It adds some helper methods along with Meta to the original Vault object.
type Vault struct {
	kong.Vault `yaml:",inline"`
	Meta
}

// Identifier returns the endpoint key name or ID.
func (v1 *Vault) Identifier() string {
	if v1.Name != nil {
		return *v1.Name
	}
	return *v1.ID
}

// Console returns an entity's identity in a human
// readable string.
func (v1 *Vault) Console() string {
	return v1.FriendlyName()
}

// Equal returns true if v1 and v2 are equal.
// TODO add compare array without position
func (v1 *Vault) Equal(v2 *Vault) bool {
	return v1.EqualWithOpts(v2, false, false)
}

// EqualWithOpts returns true if v1 and v2 are equal.
// If ignoreID is set to true, IDs will be ignored while comparison.
// If ignoreTS is set to true, timestamp fields will be ignored.
func (v1 *Vault) EqualWithOpts(v2 *Vault, ignoreID, ignoreTS bool) bool {
	v1Copy := v1.Vault.DeepCopy()
	v2Copy := v2.Vault.DeepCopy()

	if len(v1Copy.Tags) == 0 {
		v1Copy.Tags = nil
	}
	if len(v2Copy.Tags) == 0 {
		v2Copy.Tags = nil
	}

	sort.Slice(v1Copy.Tags, func(i, j int) bool { return *(v1Copy.Tags[i]) < *(v1Copy.Tags[j]) })
	sort.Slice(v2Copy.Tags, func(i, j int) bool { return *(v2Copy.Tags[i]) < *(v2Copy.Tags[j]) })

	if ignoreID {
		v1Copy.ID = nil
		v2Copy.ID = nil
	}
	if ignoreTS {
		v1Copy.CreatedAt = nil
		v2Copy.CreatedAt = nil

		v1Copy.UpdatedAt = nil
		v2Copy.UpdatedAt = nil
	}
	return reflect.DeepEqual(v1Copy, v2Copy)
}
