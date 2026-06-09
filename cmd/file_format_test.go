package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getSlice(t *testing.T, m map[string]interface{}, key string) []interface{} {
	t.Helper()
	v, ok := m[key].([]interface{})
	require.True(t, ok, "expected %q to be []interface{}", key)
	return v
}

func asMap(t *testing.T, v interface{}) map[string]interface{} {
	t.Helper()
	m, ok := v.(map[string]interface{})
	require.True(t, ok, "expected map[string]interface{}, got %T", v)
	return m
}

func TestConvertDeckToDBless_ConsumerGroups(t *testing.T) {
	input := map[string]interface{}{
		"consumer_groups": []interface{}{
			map[string]interface{}{
				"name": "group-a",
				"plugins": []interface{}{
					map[string]interface{}{"name": "rate-limiting"},
				},
			},
		},
		"consumers": []interface{}{
			map[string]interface{}{
				"username": "alice",
				"groups": []interface{}{
					map[string]interface{}{"name": "group-a"},
				},
			},
		},
	}

	result, err := convertDeckToDBless(input)
	require.NoError(t, err)

	cgPlugins := getSlice(t, result, "consumer_group_plugins")
	require.Len(t, cgPlugins, 1)
	plugin := asMap(t, cgPlugins[0])
	assert.Equal(t, "group-a", plugin["consumer_group"])
	assert.Equal(t, "rate-limiting", plugin["name"])

	cgConsumers := getSlice(t, result, "consumer_group_consumers")
	require.Len(t, cgConsumers, 1)
	entry := asMap(t, cgConsumers[0])
	assert.Equal(t, "alice", entry["consumer"])
	assert.Equal(t, "group-a", entry["consumer_group"])
}

func TestConvertDeckToDBless_ConsumerGroupPluginsRemovedFromGroup(t *testing.T) {
	input := map[string]interface{}{
		"consumer_groups": []interface{}{
			map[string]interface{}{
				"name": "grp",
				"plugins": []interface{}{
					map[string]interface{}{"name": "acl"},
				},
			},
		},
	}

	result, err := convertDeckToDBless(input)
	require.NoError(t, err)

	groups := result["consumer_groups"].([]interface{})
	group := asMap(t, groups[0])
	assert.Nil(t, group["plugins"])
}

func TestConvertDeckToDBless_NoConsumerGroups(t *testing.T) {
	input := map[string]interface{}{
		"services": []interface{}{
			map[string]interface{}{"name": "svc"},
		},
	}

	result, err := convertDeckToDBless(input)
	require.NoError(t, err)
	assert.Nil(t, result["consumer_group_plugins"])
	assert.Nil(t, result["consumer_group_consumers"])
	assert.Nil(t, result["plugins_partials"])
}

func TestConvertDeckToDBless_PluginsPartials_ByID(t *testing.T) {
	input := map[string]interface{}{
		"plugins": []interface{}{
			map[string]interface{}{
				"id": "plugin-uuid",
				"partials": []interface{}{
					map[string]interface{}{
						"id":   "partial-uuid",
						"path": "config.limit",
					},
				},
			},
		},
	}

	result, err := convertDeckToDBless(input)
	require.NoError(t, err)

	pp := getSlice(t, result, "plugins_partials")
	require.Len(t, pp, 1)
	entry := asMap(t, pp[0])
	assert.Equal(t, "plugin-uuid", entry["plugin"])
	assert.Equal(t, "partial-uuid", entry["partial"])
	assert.Equal(t, "config.limit", entry["path"])
}

func TestConvertDeckToDBless_PluginsPartials_ByName(t *testing.T) {
	input := map[string]interface{}{
		"plugins": []interface{}{
			map[string]interface{}{
				"name": "rate-limiting",
				"partials": []interface{}{
					map[string]interface{}{
						"name": "my-partial",
					},
				},
			},
		},
	}

	result, err := convertDeckToDBless(input)
	require.NoError(t, err)

	pp := getSlice(t, result, "plugins_partials")
	require.Len(t, pp, 1)
	entry := asMap(t, pp[0])
	assert.Equal(t, "rate-limiting", entry["plugin"])
	assert.Equal(t, "my-partial", entry["partial"])
}

func TestConvertDeckToDBless_PluginWithoutIDOrName(t *testing.T) {
	input := map[string]interface{}{
		"plugins": []interface{}{
			map[string]interface{}{
				"partials": []interface{}{
					map[string]interface{}{"name": "p"},
				},
			},
		},
	}

	_, err := convertDeckToDBless(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugins[0].id")
}

func TestConvertDeckToDBless_PartialWithoutIDOrName(t *testing.T) {
	input := map[string]interface{}{
		"plugins": []interface{}{
			map[string]interface{}{
				"name": "rate-limiting",
				"partials": []interface{}{
					map[string]interface{}{"path": "config.limit"},
				},
			},
		},
	}

	_, err := convertDeckToDBless(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugins[0].partials[0]")
}

func TestConvertDeckToDBless_ConsumerGroupMissingName(t *testing.T) {
	input := map[string]interface{}{
		"consumer_groups": []interface{}{
			map[string]interface{}{
				"plugins": []interface{}{
					map[string]interface{}{"name": "acl"},
				},
			},
		},
	}

	_, err := convertDeckToDBless(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "consumer_groups[0].name")
}

func TestConvertDeckToDBless_ConsumerMissingUsername(t *testing.T) {
	input := map[string]interface{}{
		"consumers": []interface{}{
			map[string]interface{}{
				"groups": []interface{}{
					map[string]interface{}{"name": "grp"},
				},
			},
		},
	}

	_, err := convertDeckToDBless(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "consumers[0].username")
}

func TestConvertDeckToDBless_GroupMissingName(t *testing.T) {
	input := map[string]interface{}{
		"consumers": []interface{}{
			map[string]interface{}{
				"username": "alice",
				"groups": []interface{}{
					map[string]interface{}{},
				},
			},
		},
	}

	_, err := convertDeckToDBless(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "consumers[0].groups[0].name")
}

func TestConvertDBlessToDeck_ConsumerGroups(t *testing.T) {
	input := map[string]interface{}{
		"_format_version": "3.0",
		"consumer_groups": []interface{}{
			map[string]interface{}{"name": "group-a"},
		},
		"consumer_group_plugins": []interface{}{
			map[string]interface{}{
				"consumer_group": "group-a",
				"name":           "rate-limiting",
			},
		},
		"consumer_group_consumers": []interface{}{
			map[string]interface{}{
				"consumer":       "alice",
				"consumer_group": "group-a",
			},
		},
		"consumers": []interface{}{
			map[string]interface{}{"username": "alice"},
		},
	}

	result, err := convertDBlessToDeck(input)
	require.NoError(t, err)

	assert.Nil(t, result["consumer_group_plugins"])
	assert.Nil(t, result["consumer_group_consumers"])

	groups := result["consumer_groups"].([]interface{})
	require.Len(t, groups, 1)
	group := asMap(t, groups[0])
	plugins := group["plugins"].([]interface{})
	require.Len(t, plugins, 1)

	consumers := result["consumers"].([]interface{})
	consumer := asMap(t, consumers[0])
	consumerGroups := consumer["groups"].([]interface{})
	require.Len(t, consumerGroups, 1)
}

func TestConvertDBlessToDeck_PluginsPartials_ByID(t *testing.T) {
	input := map[string]interface{}{
		"_format_version": "3.0",
		"plugins": []interface{}{
			map[string]interface{}{"id": "plugin-uuid", "name": "rate-limiting"},
		},
		"partials": []interface{}{
			map[string]interface{}{"id": "partial-uuid", "name": "my-partial"},
		},
		"plugins_partials": []interface{}{
			map[string]interface{}{
				"plugin":  "plugin-uuid",
				"partial": "partial-uuid",
				"path":    "config.limit",
			},
		},
	}

	result, err := convertDBlessToDeck(input)
	require.NoError(t, err)

	assert.Nil(t, result["plugins_partials"])

	plugins := result["plugins"].([]interface{})
	plugin := asMap(t, plugins[0])
	partials := plugin["partials"].([]interface{})
	require.Len(t, partials, 1)
	partial := asMap(t, partials[0])
	assert.Equal(t, "partial-uuid", partial["id"])
	assert.Equal(t, "my-partial", partial["name"])
	assert.Equal(t, "config.limit", partial["path"])
}

func TestConvertDBlessToDeck_PluginsPartials_ByName(t *testing.T) {
	input := map[string]interface{}{
		"_format_version": "3.0",
		"plugins": []interface{}{
			map[string]interface{}{"name": "rate-limiting"},
		},
		"partials": []interface{}{
			map[string]interface{}{"name": "my-partial"},
		},
		"plugins_partials": []interface{}{
			map[string]interface{}{
				"plugin":  "rate-limiting",
				"partial": "my-partial",
			},
		},
	}

	result, err := convertDBlessToDeck(input)
	require.NoError(t, err)

	plugins := result["plugins"].([]interface{})
	plugin := asMap(t, plugins[0])
	partials := plugin["partials"].([]interface{})
	require.Len(t, partials, 1)
	partial := asMap(t, partials[0])
	assert.Equal(t, "my-partial", partial["name"])
}

func TestConvertDBlessToDeck_PluginsPartials_UnresolvedPartialFallsBackToID(t *testing.T) {
	input := map[string]interface{}{
		"_format_version": "3.0",
		"plugins": []interface{}{
			map[string]interface{}{"name": "rate-limiting"},
		},
		"plugins_partials": []interface{}{
			map[string]interface{}{
				"plugin":  "rate-limiting",
				"partial": "unknown-partial-id",
			},
		},
	}

	result, err := convertDBlessToDeck(input)
	require.NoError(t, err)

	plugins := result["plugins"].([]interface{})
	plugin := asMap(t, plugins[0])
	partials := plugin["partials"].([]interface{})
	partial := asMap(t, partials[0])
	assert.Equal(t, "unknown-partial-id", partial["id"])
}

func TestConvertDBlessToDeck_PluginsPartials_MissingPlugin(t *testing.T) {
	input := map[string]interface{}{
		"_format_version": "3.0",
		"plugins": []interface{}{
			map[string]interface{}{"name": "rate-limiting"},
		},
		"plugins_partials": []interface{}{
			map[string]interface{}{
				"plugin":  "nonexistent",
				"partial": "p",
			},
		},
	}

	_, err := convertDBlessToDeck(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugins_partials[0].plugin")
}

func TestConvertDBlessToDeck_NoPluginsPartials(t *testing.T) {
	input := map[string]interface{}{
		"_format_version": "3.0",
		"consumers": []interface{}{
			map[string]interface{}{"username": "alice"},
		},
	}

	result, err := convertDBlessToDeck(input)
	require.NoError(t, err)
	assert.Nil(t, result["plugins_partials"])
}

func TestConvertDBlessToDeck_PluginsPartials_MissingPartialRef(t *testing.T) {
	input := map[string]interface{}{
		"_format_version": "3.0",
		"plugins": []interface{}{
			map[string]interface{}{"name": "rate-limiting"},
		},
		"plugins_partials": []interface{}{
			map[string]interface{}{
				"plugin": "rate-limiting",
				// "partial" key is intentionally absent
			},
		},
	}

	_, err := convertDBlessToDeck(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugins_partials[0].partial")
}

func TestConvertDeckToDBless_InvalidConsumerGroupsField(t *testing.T) {
	input := map[string]interface{}{
		"consumer_groups": "not-an-array",
	}

	_, err := convertDeckToDBless(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "consumer_groups")
}

func TestConvertDeckToDBless_InvalidConsumerGroupPluginsField(t *testing.T) {
	input := map[string]interface{}{
		"consumer_groups": []interface{}{
			map[string]interface{}{
				"name":    "grp",
				"plugins": "not-an-array",
			},
		},
	}

	_, err := convertDeckToDBless(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "consumer_groups[0].plugins")
}

func TestConvertDeckToDBless_InvalidConsumersField(t *testing.T) {
	input := map[string]interface{}{
		"consumers": "not-an-array",
	}

	_, err := convertDeckToDBless(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "consumers")
}

func TestConvertDeckToDBless_InvalidConsumerGroupsOnConsumer(t *testing.T) {
	input := map[string]interface{}{
		"consumers": []interface{}{
			map[string]interface{}{
				"username": "alice",
				"groups":   "not-an-array",
			},
		},
	}

	_, err := convertDeckToDBless(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "consumers[0].groups")
}

func TestConvertDeckToDBless_InvalidPluginsField(t *testing.T) {
	input := map[string]interface{}{
		"plugins": "not-an-array",
	}

	_, err := convertDeckToDBless(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugins")
}

func TestConvertDeckToDBless_InvalidPartialsOnPlugin(t *testing.T) {
	input := map[string]interface{}{
		"plugins": []interface{}{
			map[string]interface{}{
				"name":     "rate-limiting",
				"partials": "not-an-array",
			},
		},
	}

	_, err := convertDeckToDBless(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugins[0].partials")
}
