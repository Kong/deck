package validate

import (
	"testing"

	"github.com/kong/go-database-reconciler/pkg/state"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConflictErrors(t *testing.T) {
	plugin := func(id, name, instanceName string, global bool) *state.Plugin {
		p := &state.Plugin{Plugin: kong.Plugin{ID: kong.String(id), Name: kong.String(name)}}
		if instanceName != "" {
			p.InstanceName = kong.String(instanceName)
		}
		if !global {
			p.Service = &kong.Service{ID: kong.String("s-" + id)}
		}
		return p
	}

	t.Run("a conflict between bulk entities is not reported", func(t *testing.T) {
		errs := conflictErrors([]conflictEntity{
			{kind: "plugins", id: "p1", obj: plugin("p1", "cors", "a", false), bulk: true},
			{kind: "plugins", id: "p2", obj: plugin("p2", "cors", "a", false), bulk: true},
		})
		assert.Empty(t, errs)
	})

	t.Run("global plugins with the same name", func(t *testing.T) {
		errs := conflictErrors([]conflictEntity{
			{kind: "plugins", id: "p1", obj: plugin("p1", "cors", "", true)},
			{kind: "plugins", id: "p2", obj: plugin("p2", "cors", "", true)},
			{kind: "plugins", id: "p3", obj: plugin("p3", "cors", "", false)},
		})
		require.Len(t, errs, 1)
		require.EqualError(t, errs[0],
			`validate entity 'plugins (p2)': name of a global plugin "cors" is also used by plugins p1`)
	})

	t.Run("each unique field of an entity", func(t *testing.T) {
		consumer := func(id, username, customID string) *state.Consumer {
			return &state.Consumer{Consumer: kong.Consumer{
				ID: kong.String(id), Username: kong.String(username), CustomID: kong.String(customID),
			}}
		}
		errs := conflictErrors([]conflictEntity{
			{kind: "consumers", id: "c1", obj: consumer("c1", "user-a", "custom-a")},
			{kind: "consumers", id: "c2", obj: consumer("c2", "user-b", "custom-a")},
			{kind: "consumers", id: "c3", obj: consumer("c3", "user-a", "custom-c")},
		})
		require.Len(t, errs, 2)
		require.EqualError(t, errs[0], `validate entity 'consumers (c2)': custom_id "custom-a" is also used by consumers c1`)
		require.EqualError(t, errs[1], `validate entity 'consumers (c3)': username "user-a" is also used by consumers c1`)
	})
}
