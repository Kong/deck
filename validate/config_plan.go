package validate

import (
	"encoding/json"
	"fmt"

	"github.com/kong/go-database-reconciler/pkg/state"
	"github.com/kong/go-kong/kong"
)

// configKind is an entity list that the config validation endpoint accepts.
type configKind struct {
	// name is the key of the list in the request body.
	name string
	// stateField is the EntityMap key of the list. It is empty if --online-entities-list cannot select the list.
	stateField string
	// perEntity is true if the list has a /schemas/<name>/validate endpoint.
	perEntity bool
	// credential is true for consumer credentials. Older versions of the endpoint do not accept them.
	credential bool
}

// configKinds puts each referenced kind before the kinds that refer to it.
var configKinds = []configKind{
	{name: "certificates", stateField: "Certificates", perEntity: true},
	{name: "ca_certificates", stateField: "CACertificates", perEntity: true},
	{name: "key_sets", perEntity: true},
	{name: "keys", perEntity: true},
	{name: "consumer_groups", perEntity: true},
	{name: "consumers", stateField: "Consumers", perEntity: true},
	{name: "keyauth_credentials", stateField: "KeyAuths", perEntity: true, credential: true},
	{name: "basicauth_credentials", stateField: "BasicAuths", perEntity: true, credential: true},
	{name: "hmacauth_credentials", stateField: "HMACAuths", perEntity: true, credential: true},
	{name: "jwt_secrets", stateField: "JWTAuths", perEntity: true, credential: true},
	{name: "acls", stateField: "ACLGroups", perEntity: true, credential: true},
	{name: "mtls_auth_credentials", perEntity: true, credential: true},
	{name: "services", stateField: "Services", perEntity: true},
	{name: "routes", stateField: "Routes", perEntity: true},
	{name: "upstreams", stateField: "Upstreams", perEntity: true},
	{name: "targets", stateField: "Targets", perEntity: true},
	{name: "snis", stateField: "SNIs", perEntity: true},
	{name: "vaults", stateField: "Vaults", perEntity: true},
	{name: "cloned_plugins"},
	{name: "plugins", stateField: "Plugins", perEntity: true},
}

// configKindOfField returns the entity list of an EntityMap key.
func configKindOfField(field string) (configKind, bool) {
	for _, kind := range configKinds {
		if field != "" && kind.stateField == field {
			return kind, true
		}
	}
	return configKind{}, false
}

func supportsPerEntityValidation(name string) bool {
	for _, kind := range configKinds {
		if kind.name == name {
			return kind.perEntity
		}
	}
	return false
}

type configPlan struct {
	// entities holds the entities to validate with the config validation endpoint.
	entities []*configEntity
	// perEntityPlugins holds the plugins that the config validation endpoint does not support.
	perEntityPlugins []any
}

type configIndex map[string]map[string]*configEntity

func (idx configIndex) get(kind string, id *string) *configEntity {
	if id == nil {
		return nil
	}
	return idx[kind][*id]
}

func (idx configIndex) add(kind string, id *string, obj, kongObj any) (*configEntity, error) {
	if id == nil || *id == "" {
		return nil, fmt.Errorf("%s entity without an ID", kind)
	}
	raw, err := json.Marshal(kongObj)
	if err != nil {
		return nil, fmt.Errorf("encoding %s %s: %w", kind, *id, err)
	}
	if idx[kind] == nil {
		idx[kind] = map[string]*configEntity{}
	}
	e := &configEntity{kind: kind, id: *id, obj: obj, raw: raw}
	idx[kind][*id] = e
	return e, nil
}

func addDeps(e *configEntity, deps ...*configEntity) {
	for _, d := range deps {
		if d != nil {
			e.deps = append(e.deps, d)
		}
	}
}

func isCustomPlugin(p *state.Plugin, customPlugins map[string]bool) bool {
	return p.Name != nil && customPlugins[*p.Name]
}

// buildConfigPlan selects the entities of kinds for the config validation endpoint.
// A plugin with a custom plugin name or with partials gets the per-entity validation.
func buildConfigPlan(ks *state.KongState, kinds map[string]bool, customPlugins map[string]bool) (*configPlan, error) {
	idx := configIndex{}
	var all []*configEntity
	add := func(kind string, id *string, obj, kongObj any) (*configEntity, error) {
		e, err := idx.add(kind, id, obj, kongObj)
		if err == nil {
			all = append(all, e)
		}
		return e, err
	}

	certificates, err := ks.Certificates.GetAll()
	if err != nil {
		return nil, err
	}
	for _, c := range certificates {
		if _, err := add("certificates", c.ID, c, &c.Certificate); err != nil {
			return nil, err
		}
	}
	caCertificates, err := ks.CACertificates.GetAll()
	if err != nil {
		return nil, err
	}
	for _, c := range caCertificates {
		if _, err := add("ca_certificates", c.ID, c, &c.CACertificate); err != nil {
			return nil, err
		}
	}
	keySets, err := ks.KeySets.GetAll()
	if err != nil {
		return nil, err
	}
	for _, s := range keySets {
		if _, err := add("key_sets", s.ID, s, &s.KeySet); err != nil {
			return nil, err
		}
	}
	keys, err := ks.Keys.GetAll()
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		e, err := add("keys", k.ID, k, &k.Key)
		if err != nil {
			return nil, err
		}
		if k.Set != nil {
			addDeps(e, idx.get("key_sets", k.Set.ID))
		}
	}
	consumerGroups, err := ks.ConsumerGroups.GetAll()
	if err != nil {
		return nil, err
	}
	for _, cg := range consumerGroups {
		if _, err := add("consumer_groups", cg.ID, cg, &cg.ConsumerGroup); err != nil {
			return nil, err
		}
	}
	consumers, err := ks.Consumers.GetAll()
	if err != nil {
		return nil, err
	}
	for _, c := range consumers {
		if _, err := add("consumers", c.ID, c, &c.Consumer); err != nil {
			return nil, err
		}
	}
	services, err := ks.Services.GetAll()
	if err != nil {
		return nil, err
	}
	for _, s := range services {
		e, err := add("services", s.ID, s, &s.Service)
		if err != nil {
			return nil, err
		}
		if s.ClientCertificate != nil {
			addDeps(e, idx.get("certificates", s.ClientCertificate.ID))
		}
		for _, id := range s.CACertificates {
			addDeps(e, idx.get("ca_certificates", id))
		}
	}
	routes, err := ks.Routes.GetAll()
	if err != nil {
		return nil, err
	}
	for _, r := range routes {
		e, err := add("routes", r.ID, r, &r.Route)
		if err != nil {
			return nil, err
		}
		if r.Service != nil {
			addDeps(e, idx.get("services", r.Service.ID))
		}
	}
	upstreams, err := ks.Upstreams.GetAll()
	if err != nil {
		return nil, err
	}
	for _, u := range upstreams {
		e, err := add("upstreams", u.ID, u, &u.Upstream)
		if err != nil {
			return nil, err
		}
		if u.ClientCertificate != nil {
			addDeps(e, idx.get("certificates", u.ClientCertificate.ID))
		}
	}
	targets, err := ks.Targets.GetAll()
	if err != nil {
		return nil, err
	}
	for _, t := range targets {
		e, err := add("targets", t.ID, t, &t.Target)
		if err != nil {
			return nil, err
		}
		if t.Upstream != nil {
			addDeps(e, idx.get("upstreams", t.Upstream.ID))
		}
	}
	snis, err := ks.SNIs.GetAll()
	if err != nil {
		return nil, err
	}
	for _, s := range snis {
		e, err := add("snis", s.ID, s, &s.SNI)
		if err != nil {
			return nil, err
		}
		if s.Certificate != nil {
			addDeps(e, idx.get("certificates", s.Certificate.ID))
		}
	}
	vaults, err := ks.Vaults.GetAll()
	if err != nil {
		return nil, err
	}
	for _, v := range vaults {
		if _, err := add("vaults", v.ID, v, &v.Vault); err != nil {
			return nil, err
		}
	}
	credentials, err := credentialEntries(ks)
	if err != nil {
		return nil, err
	}
	for _, c := range credentials {
		e, err := add(c.kind, c.id, c.obj, c.kongObj)
		if err != nil {
			return nil, err
		}
		if c.consumer != nil {
			addDeps(e, idx.get("consumers", c.consumer.ID))
		}
		if c.caCertificate != nil {
			addDeps(e, idx.get("ca_certificates", c.caCertificate.ID))
		}
	}

	clonedPlugins, err := ks.ClonedPluginDefinitions.GetAll()
	if err != nil {
		return nil, err
	}
	clonedByName := map[string]*configEntity{}
	for _, c := range clonedPlugins {
		e, err := add("cloned_plugins", c.ID, c, &c.ClonedPluginDefinition)
		if err != nil {
			return nil, err
		}
		if c.Name != nil {
			clonedByName[*c.Name] = e
		}
	}

	plan := &configPlan{}
	plugins, err := ks.Plugins.GetAll()
	if err != nil {
		return nil, err
	}
	for _, p := range plugins {
		if isCustomPlugin(p, customPlugins) || len(p.Partials) > 0 {
			if kinds["plugins"] {
				plan.perEntityPlugins = append(plan.perEntityPlugins, p)
			}
			continue
		}
		e, err := add("plugins", p.ID, p, &p.Plugin)
		if err != nil {
			return nil, err
		}
		if p.Route != nil {
			addDeps(e, idx.get("routes", p.Route.ID))
		}
		if p.Service != nil {
			addDeps(e, idx.get("services", p.Service.ID))
		}
		if p.Consumer != nil {
			addDeps(e, idx.get("consumers", p.Consumer.ID))
		}
		if p.ConsumerGroup != nil {
			addDeps(e, idx.get("consumer_groups", p.ConsumerGroup.ID))
		}
		if p.Name != nil {
			addDeps(e, clonedByName[*p.Name])
		}
	}

	for _, e := range orderConfigEntities(all) {
		if kinds[e.kind] {
			plan.entities = append(plan.entities, e)
		}
	}
	return plan, nil
}

// conflictEntity is an entity for the local conflict check.
type conflictEntity struct {
	kind string
	id   string
	obj  any
	// bulk is true if the config validation endpoint gets the entity.
	bulk bool
}

// conflictErrors returns an error for each entity that has a conflict key of an entity before it.
// A conflict between two bulk entities is not an error here. The config validation endpoint reports it.
func conflictErrors(entities []conflictEntity) []error {
	var errs []error
	owners := map[conflictKey]conflictEntity{}
	for _, e := range entities {
		for _, key := range conflictKeys(e.obj) {
			owner, ok := owners[key]
			if !ok {
				owners[key] = e
				continue
			}
			if owner.bulk && e.bulk {
				continue
			}
			errs = append(errs, fmt.Errorf("validate entity '%s (%s)': %s %q is also used by %s %s",
				e.kind, e.id, key.field, key.value, owner.kind, owner.id))
		}
	}
	return errs
}

// conflictEntities returns the entities of plan for the local conflict check.
func (p *configPlan) conflictEntities(bulk bool) []conflictEntity {
	entities := make([]conflictEntity, 0, len(p.entities)+len(p.perEntityPlugins))
	for _, e := range p.entities {
		entities = append(entities, conflictEntity{kind: e.kind, id: e.id, obj: e.obj, bulk: bulk})
	}
	for _, obj := range p.perEntityPlugins {
		if p, ok := obj.(*state.Plugin); ok && p.ID != nil {
			entities = append(entities, conflictEntity{kind: "plugins", id: *p.ID, obj: p})
		}
	}
	return entities
}

// credentialEntry is a consumer credential of the state.
type credentialEntry struct {
	kind          string
	id            *string
	obj           any
	kongObj       any
	consumer      *kong.Consumer
	caCertificate *kong.CACertificate
}

type credentialCollector struct {
	entries []credentialEntry
	err     error
}

func collectCredentials[T any](c *credentialCollector, getAll func() ([]*T, error), entry func(*T) credentialEntry) {
	if c.err != nil {
		return
	}
	items, err := getAll()
	if err != nil {
		c.err = err
		return
	}
	for _, item := range items {
		c.entries = append(c.entries, entry(item))
	}
}

func credentialEntries(ks *state.KongState) ([]credentialEntry, error) {
	c := &credentialCollector{}
	collectCredentials(c, ks.KeyAuths.GetAll, func(k *state.KeyAuth) credentialEntry {
		return credentialEntry{kind: "keyauth_credentials", id: k.ID, obj: k, kongObj: &k.KeyAuth, consumer: k.Consumer}
	})
	collectCredentials(c, ks.BasicAuths.GetAll, func(b *state.BasicAuth) credentialEntry {
		return credentialEntry{
			kind: "basicauth_credentials", id: b.ID, obj: b, kongObj: &b.BasicAuth, consumer: b.Consumer,
		}
	})
	collectCredentials(c, ks.HMACAuths.GetAll, func(h *state.HMACAuth) credentialEntry {
		return credentialEntry{kind: "hmacauth_credentials", id: h.ID, obj: h, kongObj: &h.HMACAuth, consumer: h.Consumer}
	})
	collectCredentials(c, ks.JWTAuths.GetAll, func(j *state.JWTAuth) credentialEntry {
		return credentialEntry{kind: "jwt_secrets", id: j.ID, obj: j, kongObj: &j.JWTAuth, consumer: j.Consumer}
	})
	collectCredentials(c, ks.ACLGroups.GetAll, func(a *state.ACLGroup) credentialEntry {
		return credentialEntry{kind: "acls", id: a.ID, obj: a, kongObj: &a.ACLGroup, consumer: a.Consumer}
	})
	collectCredentials(c, ks.MTLSAuths.GetAll, func(m *state.MTLSAuth) credentialEntry {
		return credentialEntry{
			kind: "mtls_auth_credentials", id: m.ID, obj: m, kongObj: &m.MTLSAuth,
			consumer: m.Consumer, caCertificate: m.CACertificate,
		}
	})
	return c.entries, c.err
}
