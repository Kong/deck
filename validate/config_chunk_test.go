package validate

import (
	"strings"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/state"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunkConfig(t *testing.T) {
	entity := func(kind, id string, size int, deps ...*configEntity) *configEntity {
		return &configEntity{kind: kind, id: id, raw: []byte(strings.Repeat("x", size)), deps: deps}
	}
	limit := configEnvelopeSize() + 100

	t.Run("everything fits in one chunk", func(t *testing.T) {
		svc := entity("services", "s1", 10)
		route := entity("routes", "r1", 10, svc)
		chunks := chunkConfig([]*configEntity{svc, route}, limit)
		require.Len(t, chunks, 1)
		assert.Equal(t, []*configEntity{svc, route}, chunks[0].entities)
	})

	t.Run("a reference goes into each chunk that needs it", func(t *testing.T) {
		svc := entity("services", "s1", 30)
		r1 := entity("routes", "r1", 30, svc)
		r2 := entity("routes", "r2", 30, svc)
		r3 := entity("routes", "r3", 30, svc)
		chunks := chunkConfig([]*configEntity{svc, r1, r2, r3}, limit)
		require.Len(t, chunks, 2)
		assert.Equal(t, []*configEntity{svc, r1, r2}, chunks[0].entities)
		assert.Equal(t, []*configEntity{svc, r3}, chunks[1].entities)
		assert.Equal(t, []*configEntity{r3}, chunks[1].primary)
	})

	t.Run("an entity larger than the limit gets its own chunk", func(t *testing.T) {
		small := entity("services", "s1", 10)
		large := entity("services", "s2", 500)
		chunks := chunkConfig([]*configEntity{small, large, small}, limit)
		require.Len(t, chunks, 3)
		assert.Equal(t, []*configEntity{large}, chunks[1].entities)
	})

	t.Run("body groups entities by kind", func(t *testing.T) {
		svc := &configEntity{kind: "services", id: "s1", raw: []byte(`{"id":"s1"}`)}
		route := &configEntity{kind: "routes", id: "r1", raw: []byte(`{"id":"r1"}`), deps: []*configEntity{svc}}
		plugin := &configEntity{kind: "plugins", id: "p1", raw: []byte(`{"id":"p1"}`), deps: []*configEntity{route}}
		chunks := chunkConfig([]*configEntity{plugin}, maxConfigChunkBytes)
		require.Len(t, chunks, 1)
		assert.JSONEq(t,
			`{"services":[{"id":"s1"}],"routes":[{"id":"r1"}],"plugins":[{"id":"p1"}]}`,
			string(chunks[0].body()))
	})
}

func TestOrderConfigEntities(t *testing.T) {
	s1 := &configEntity{kind: "services", id: "s1"}
	s2 := &configEntity{kind: "services", id: "s2"}
	r1 := &configEntity{kind: "routes", id: "r1", deps: []*configEntity{s1}}
	r2 := &configEntity{kind: "routes", id: "r2", deps: []*configEntity{s2}}
	p1 := &configEntity{kind: "plugins", id: "p1", deps: []*configEntity{r1}}
	p2 := &configEntity{kind: "plugins", id: "p2", deps: []*configEntity{s2}}

	ordered := orderConfigEntities([]*configEntity{s1, s2, r1, r2, p1, p2})

	assert.Equal(t, []*configEntity{s1, r1, p1, s2, r2, p2}, ordered)
}

func TestConflictGroups(t *testing.T) {
	plugin := func(instanceName string) *configEntity {
		return &configEntity{kind: "plugins", obj: &state.Plugin{Plugin: kong.Plugin{
			Name: kong.String("cors"), InstanceName: kong.String(instanceName), Route: &kong.Route{},
		}}}
	}
	a1, b1, c1, a2, b2 := plugin("a"), plugin("b"), plugin("c"), plugin("a"), plugin("b")

	groups := conflictGroups([]*configEntity{a1, b1, c1, a2, b2})

	assert.Equal(t, [][]*configEntity{{a1, a2}, {b1, b2}, {c1}}, groups)
}
