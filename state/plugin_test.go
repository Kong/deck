package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func TestPluginInsert(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewPluginsCollection()
	assert.Nil(err)
	assert.NotNil(collection)
	var plugin Plugin
	plugin.Name = kong.String("my-plugin")
	plugin.ID = kong.String("first")
	err = collection.Add(plugin)
	assert.Nil(err)
}

func TestPluginGetUpdate(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewPluginsCollection()
	assert.Nil(err)
	assert.NotNil(collection)
	var plugin Plugin
	plugin.Name = kong.String("my-plugin")
	plugin.ID = kong.String("first")
	plugin.Service = &kong.Service{
		ID:   kong.String("service1-id"),
		Name: kong.String("service1-name"),
	}
	assert.NotNil(plugin.Service)
	err = collection.Add(plugin)
	assert.NotNil(plugin.Service)
	assert.Nil(err)

	re, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Equal("my-plugin", *re.Name)
	re.Service = &kong.Service{
		ID:   kong.String("service2-id"),
		Name: kong.String("service2-name"),
	}
	err = collection.Update(*re)
	assert.Nil(err)

	re, err = collection.Get("my-plugin")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Equal("service2-id", *re.Service.ID)

	re, err = collection.Get("does-not-exists")
	assert.Equal(ErrNotFound, err)
	assert.Nil(re)
}

func TestGetPluginByProp(t *testing.T) {
	plugins := []Plugin{
		{
			Plugin: kong.Plugin{
				ID:   kong.String("1"),
				Name: kong.String("key-auth"),
				Config: map[string]interface{}{
					"key1": "value1",
				},
			},
		},
		{
			Plugin: kong.Plugin{
				ID:   kong.String("2"),
				Name: kong.String("key-auth"),
				Service: &kong.Service{
					Name: kong.String("svc1"),
				},
				Config: map[string]interface{}{
					"key2": "value2",
				},
			},
		},
		{
			Plugin: kong.Plugin{
				ID:   kong.String("3"),
				Name: kong.String("key-auth"),
				Route: &kong.Route{
					Name: kong.String("route1"),
				},
				Config: map[string]interface{}{
					"key3": "value3",
				},
			},
		},
		{
			Plugin: kong.Plugin{
				ID:   kong.String("4"),
				Name: kong.String("key-auth"),
				Consumer: &kong.Consumer{
					Username: kong.String("consumer1"),
				},
				Config: map[string]interface{}{
					"key4": "value4",
				},
			},
		},
	}
	assert := assert.New(t)
	collection, err := NewPluginsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	for _, p := range plugins {
		assert.Nil(collection.Add(p))
	}

	plugin, err := collection.GetByProp("foo", "", "", "")
	assert.Nil(plugin)
	assert.Equal(ErrNotFound, err)

	plugin, err = collection.GetByProp("key-auth", "", "", "")
	assert.Nil(err)
	assert.NotNil(plugin)
	assert.Equal("value1", plugin.Config["key1"])

	plugin, err = collection.GetByProp("key-auth", "svc1", "", "")
	assert.Nil(err)
	assert.NotNil(plugin)
	assert.Equal("value2", plugin.Config["key2"])

	plugin, err = collection.GetByProp("key-auth", "", "route1", "")
	assert.Nil(err)
	assert.NotNil(plugin)
	assert.Equal("value3", plugin.Config["key3"])

	plugin, err = collection.GetByProp("key-auth", "", "", "consumer1")
	assert.Nil(err)
	assert.NotNil(plugin)
	assert.Equal("value4", plugin.Config["key4"])
}

func TestPluginsInvalidType(t *testing.T) {
	assert := assert.New(t)

	collection, err := NewPluginsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var service Service
	service.Name = kong.String("my-service")
	service.ID = kong.String("first")
	txn := collection.memdb.Txn(true)
	txn.Insert(pluginTableName, &service)
	txn.Commit()

	assert.Panics(func() {
		collection.Get("my-service")
	})
}

func TestPluginDelete(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewPluginsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var plugin Plugin
	plugin.ID = kong.String("first")
	plugin.Name = kong.String("my-plugin")
	plugin.Config = map[string]interface{}{
		"foo": "bar",
		"baz": "bar",
	}
	plugin.Service = &kong.Service{
		ID:   kong.String("service1-id"),
		Name: kong.String("service1-name"),
	}
	err = collection.Add(plugin)
	assert.Nil(err)

	p, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(p)
	assert.Equal("bar", p.Config["foo"])

	err = collection.Delete(*p.ID)
	assert.Nil(err)

	err = collection.Delete(*p.ID)
	assert.NotNil(err)
}

func TestPluginGetAll(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewPluginsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	plugins := []*Plugin{
		{
			Plugin: kong.Plugin{
				ID:   kong.String("first-id"),
				Name: kong.String("key-auth"),
				Service: &kong.Service{
					ID:   kong.String("service1-id"),
					Name: kong.String("service1-name"),
				},
				Config: map[string]interface{}{
					"foo": "bar",
					"baz": "bar",
				},
			},
		},
		{
			Plugin: kong.Plugin{
				ID:   kong.String("second-id"),
				Name: kong.String("basic-auth"),
				Service: &kong.Service{
					ID:   kong.String("service1-id"),
					Name: kong.String("service1-name"),
				},
			},
		},
		{
			Plugin: kong.Plugin{
				ID:   kong.String("third-id"),
				Name: kong.String("rate-limiting"),
				Route: &kong.Route{
					ID:   kong.String("route1-id"),
					Name: kong.String("route1-name"),
				},
			},
		},
		{
			Plugin: kong.Plugin{
				ID:   kong.String("fourth-id"),
				Name: kong.String("key-auth"),
				Route: &kong.Route{
					ID:   kong.String("route1-id"),
					Name: kong.String("route1-name"),
				},
			},
		},
	}

	for _, p := range plugins {
		assert.Nil(collection.Add(*p))
	}

	allPlugins, err := collection.GetAll()
	assert.Nil(err)
	assert.Equal(len(plugins), len(allPlugins))

	allPlugins, err = collection.GetAllByName("key-auth")
	assert.Nil(err)
	assert.Equal(2, len(allPlugins))

	allPlugins, err = collection.GetAllByRouteID("route1-id")
	assert.Nil(err)
	assert.Equal(2, len(allPlugins))

	allPlugins, err = collection.GetAllByServiceID("service1-id")
	assert.Nil(err)
	assert.Equal(2, len(allPlugins))

	allPlugins, err = collection.GetAllByServiceID("service-nope")
	assert.Nil(err)
	assert.Equal(0, len(allPlugins))
}
