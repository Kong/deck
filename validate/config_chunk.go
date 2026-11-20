package validate

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/state"
	"github.com/kong/go-kong/kong"
)

// configEntity is one entity of a configuration validation request.
type configEntity struct {
	kind string
	id   string
	// obj is the state entity. The per-entity validation sends it.
	obj any
	raw []byte
	// deps holds the entities that this entity refers to. The first one is the parent for chunk locality.
	deps []*configEntity
}

type configChunk struct {
	entities []*configEntity
	// primary holds the entities that the chunk validates. The other entities are references of these.
	primary []*configEntity
	members map[*configEntity]bool
	size    int
}

func newConfigChunk() *configChunk {
	return &configChunk{members: map[*configEntity]bool{}}
}

// missing returns the entities of group and their transitive references that are not in the chunk.
func (c *configChunk) missing(group []*configEntity) []*configEntity {
	var out []*configEntity
	seen := map[*configEntity]bool{}
	var visit func(*configEntity)
	visit = func(x *configEntity) {
		if seen[x] || c.members[x] {
			return
		}
		seen[x] = true
		for _, d := range x.deps {
			visit(d)
		}
		out = append(out, x)
	}
	for _, e := range group {
		visit(e)
	}
	return out
}

func entitiesSize(entities []*configEntity) int {
	size := 0
	for _, e := range entities {
		size += len(e.raw) + 1
	}
	return size
}

func (c *configChunk) add(entities []*configEntity) {
	for _, e := range entities {
		c.members[e] = true
		c.entities = append(c.entities, e)
	}
	c.size += entitiesSize(entities)
}

func (c *configChunk) body() []byte {
	byKind := map[string][]*configEntity{}
	for _, e := range c.entities {
		byKind[e.kind] = append(byKind[e.kind], e)
	}
	var buf bytes.Buffer
	buf.Grow(c.size + configEnvelopeSize())
	buf.WriteByte('{')
	first := true
	for _, kind := range configKinds {
		entities := byKind[kind.name]
		if len(entities) == 0 {
			continue
		}
		if !first {
			buf.WriteByte(',')
		}
		first = false
		fmt.Fprintf(&buf, "%q:[", kind.name)
		for i, e := range entities {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.Write(e.raw)
		}
		buf.WriteByte(']')
	}
	buf.WriteByte('}')
	return buf.Bytes()
}

// configEnvelopeSize is the maximum size of the request body without the entities.
func configEnvelopeSize() int {
	size := 2
	for _, kind := range configKinds {
		size += len(kind.name) + len(`"":[],`)
	}
	return size
}

// chunkConfig splits entities into chunks of at most limit bytes.
// Each chunk holds the entities that its entities refer to.
// Entities that can conflict with each other go into the same chunk.
// A group that is larger than limit with its references gets a chunk of its own.
func chunkConfig(entities []*configEntity, limit int) []*configChunk {
	limit -= configEnvelopeSize()
	var chunks []*configChunk
	current := newConfigChunk()
	for _, group := range conflictGroups(entities) {
		missing := current.missing(group)
		if len(current.entities) > 0 && current.size+entitiesSize(missing) > limit {
			chunks = append(chunks, current)
			current = newConfigChunk()
			missing = current.missing(group)
		}
		current.add(missing)
		current.primary = append(current.primary, group...)
	}
	if len(current.entities) > 0 {
		chunks = append(chunks, current)
	}
	return chunks
}

// conflictKey is a value that must be unique in one configuration.
type conflictKey struct {
	kind  string
	field string
	value string
}

// conflictKeys returns the unique values that the config validation endpoint checks for obj.
func conflictKeys(obj any) []conflictKey {
	var keys []conflictKey
	add := func(kind, field string, value *string) {
		if value != nil && *value != "" {
			keys = append(keys, conflictKey{kind: kind, field: field, value: *value})
		}
	}
	switch o := obj.(type) {
	case *state.Service:
		add("services", "name", o.Name)
	case *state.Route:
		add("routes", "name", o.Name)
	case *state.Upstream:
		add("upstreams", "name", o.Name)
	case *state.SNI:
		add("snis", "name", o.Name)
	case *state.ConsumerGroup:
		add("consumer_groups", "name", o.Name)
	case *state.Key:
		add("keys", "name", o.Name)
	case *state.KeySet:
		add("key_sets", "name", o.Name)
	case *state.ClonedPluginDefinition:
		add("cloned_plugins", "name", o.Name)
	case *state.Consumer:
		add("consumers", "username", o.Username)
		add("consumers", "custom_id", o.CustomID)
	case *state.CACertificate:
		if o.Cert != nil {
			add("ca_certificates", "cert", kong.String(strings.TrimSpace(*o.Cert)))
		}
	case *state.Vault:
		add("vaults", "prefix", o.Prefix)
	case *state.KeyAuth:
		add("keyauth_credentials", "key", o.Key)
	case *state.JWTAuth:
		add("jwt_secrets", "key", o.Key)
	case *state.BasicAuth:
		add("basicauth_credentials", "username", o.Username)
	case *state.HMACAuth:
		add("hmacauth_credentials", "username", o.Username)
	case *state.MTLSAuth:
		add("mtls_auth_credentials", "subject_name", o.SubjectName)
	case *state.ACLGroup:
		if o.Consumer != nil && o.Consumer.ID != nil && o.Group != nil {
			add("acls", "consumer and group", kong.String(*o.Consumer.ID+":"+*o.Group))
		}
	case *state.Plugin:
		add("plugins", "instance_name", o.InstanceName)
		if o.Service == nil && o.Route == nil && o.Consumer == nil && o.ConsumerGroup == nil {
			add("plugins", "name of a global plugin", o.Name)
		}
	}
	return keys
}

// conflictGroups puts entities that share a conflict key into one group.
// Each group is at the position of its first entity.
func conflictGroups(entities []*configEntity) [][]*configEntity {
	parent := make([]int, len(entities))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(i int) int {
		if parent[i] != i {
			parent[i] = find(parent[i])
		}
		return parent[i]
	}
	owner := map[conflictKey]int{}
	for i, e := range entities {
		for _, key := range conflictKeys(e.obj) {
			j, ok := owner[key]
			if !ok {
				owner[key] = i
				continue
			}
			ri, rj := find(i), find(j)
			parent[max(ri, rj)] = min(ri, rj)
		}
	}
	members := map[int][]*configEntity{}
	for i, e := range entities {
		root := find(i)
		members[root] = append(members[root], e)
	}
	groups := make([][]*configEntity, 0, len(members))
	for i := range entities {
		if find(i) == i {
			groups = append(groups, members[i])
		}
	}
	return groups
}

// orderConfigEntities puts each entity after its parent and next to its siblings.
// Entities that refer to each other then share a chunk more often.
func orderConfigEntities(all []*configEntity) []*configEntity {
	children := map[*configEntity][]*configEntity{}
	var roots []*configEntity
	for _, e := range all {
		if len(e.deps) == 0 {
			roots = append(roots, e)
			continue
		}
		children[e.deps[0]] = append(children[e.deps[0]], e)
	}
	ordered := make([]*configEntity, 0, len(all))
	var visit func(*configEntity)
	visit = func(e *configEntity) {
		ordered = append(ordered, e)
		for _, c := range children[e] {
			visit(c)
		}
	}
	for _, r := range roots {
		visit(r)
	}
	return ordered
}
