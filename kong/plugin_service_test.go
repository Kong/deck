package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestPluginsService(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	plugin := &Plugin{
		Name: String("key-auth"),
	}

	createdPlugin, err := client.Plugins.Create(defaultCtx, plugin)
	assert.Nil(err)
	assert.NotNil(createdPlugin)

	plugin, err = client.Plugins.Get(defaultCtx, createdPlugin.ID)
	assert.Nil(err)
	assert.NotNil(plugin)

	plugin.Config["key_in_body"] = true
	plugin, err = client.Plugins.Update(defaultCtx, plugin)
	assert.Nil(err)
	assert.NotNil(plugin)
	assert.Equal(true, plugin.Config["key_in_body"])

	err = client.Plugins.Delete(defaultCtx, createdPlugin.ID)
	assert.Nil(err)

	// ID can be specified
	id := uuid.NewV4().String()
	plugin = &Plugin{
		Name: String("prometheus"),
		ID:   String(id),
	}

	createdPlugin, err = client.Plugins.Create(defaultCtx, plugin)
	assert.Nil(err)
	assert.NotNil(createdPlugin)
	assert.Equal(id, *createdPlugin.ID)

	err = client.Plugins.Delete(defaultCtx, createdPlugin.ID)
	assert.Nil(err)
}

func TestPluginWithTags(T *testing.T) {
	runWhenKong(T, ">=1.1.0")
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	plugin := &Plugin{
		Name: String("key-auth"),
		Tags: StringSlice("tag1", "tag2"),
	}

	createdPlugin, err := client.Plugins.Create(defaultCtx, plugin)
	assert.Nil(err)
	assert.NotNil(createdPlugin)
	assert.Equal(StringSlice("tag1", "tag2"), createdPlugin.Tags)

	err = client.Plugins.Delete(defaultCtx, createdPlugin.ID)
	assert.Nil(err)
}

func TestUnknownPlugin(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	plugin, err := client.Plugins.Create(defaultCtx, &Plugin{
		Name: String("plugin-not-present"),
	})
	assert.NotNil(err)
	assert.Nil(plugin)
}

func TestPluginListEndpoint(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// fixtures
	plugins := []*Plugin{
		{
			Name: String("key-auth"),
		},
		{
			Name: String("basic-auth"),
		},
		{
			Name: String("jwt"),
		},
	}

	// create fixturs
	for i := 0; i < len(plugins); i++ {
		plugin, err := client.Plugins.Create(defaultCtx, plugins[i])
		assert.Nil(err)
		assert.NotNil(plugin)
		plugins[i] = plugin
	}

	pluginsFromKong, next, err := client.Plugins.List(defaultCtx, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(pluginsFromKong)
	assert.Equal(3, len(pluginsFromKong))

	// check if we see all plugins
	assert.True(comparePlugins(plugins, pluginsFromKong))

	// Test pagination
	pluginsFromKong = []*Plugin{}

	// first page
	page1, next, err := client.Plugins.List(defaultCtx, &ListOpt{Size: 1})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))
	pluginsFromKong = append(pluginsFromKong, page1...)

	// second page
	page2, next, err := client.Plugins.List(defaultCtx, next)
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page2)
	assert.Equal(1, len(page2))
	pluginsFromKong = append(pluginsFromKong, page2...)

	// last page
	page3, next, err := client.Plugins.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(page3)
	assert.Equal(1, len(page3))
	pluginsFromKong = append(pluginsFromKong, page3...)

	assert.True(comparePlugins(plugins, pluginsFromKong))

	plugins, err = client.Plugins.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(plugins)
	assert.Equal(3, len(plugins))

	for i := 0; i < len(plugins); i++ {
		assert.Nil(client.Plugins.Delete(defaultCtx, plugins[i].ID))
	}
}

func TestPluginListAllForEntityEndpoint(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// fixtures

	createdService, err := client.Services.Create(defaultCtx, &Service{
		Name: String("foo"),
		Host: String("upstream"),
		Port: Int(42),
		Path: String("/path"),
	})
	assert.Nil(err)
	assert.NotNil(createdService)

	createdRoute, err := client.Routes.Create(defaultCtx, &Route{
		Hosts:   StringSlice("host1.com", "host2.com"),
		Service: createdService,
	})
	assert.Nil(err)
	assert.NotNil(createdRoute)

	createdConsumer, err := client.Consumers.Create(defaultCtx, &Consumer{
		Username: String("foo"),
	})
	assert.Nil(err)
	assert.NotNil(createdConsumer)

	plugins := []*Plugin{
		// global
		{
			Name: String("key-auth"),
		},
		{
			Name: String("basic-auth"),
		},
		{
			Name: String("jwt"),
		},
		// specific to route
		{
			Name:  String("key-auth"),
			Route: createdRoute,
		},
		{
			Name:  String("jwt"),
			Route: createdRoute,
		},
		// specific to service
		{
			Name:    String("key-auth"),
			Service: createdService,
		},
		{
			Name:    String("jwt"),
			Service: createdService,
		},
		// specific to consumer
		{
			Name:     String("rate-limiting"),
			Consumer: createdConsumer,
			Config: map[string]interface{}{
				"second": 1,
			},
		},
	}

	// create fixturs
	for i := 0; i < len(plugins); i++ {
		plugin, err := client.Plugins.Create(defaultCtx, plugins[i])
		assert.Nil(err)
		assert.NotNil(plugin)
		plugins[i] = plugin
	}

	pluginsFromKong, err := client.Plugins.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(pluginsFromKong)
	assert.Equal(len(plugins), len(pluginsFromKong))

	// check if we see all plugins
	assert.True(comparePlugins(plugins, pluginsFromKong))

	assert.True(comparePlugins(plugins, pluginsFromKong))

	pluginsFromKong, err = client.Plugins.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(pluginsFromKong)
	assert.Equal(8, len(pluginsFromKong))

	pluginsFromKong, err = client.Plugins.ListAllForConsumer(defaultCtx,
		createdConsumer.ID)
	assert.Nil(err)
	assert.NotNil(pluginsFromKong)
	assert.Equal(1, len(pluginsFromKong))

	pluginsFromKong, err = client.Plugins.ListAllForService(defaultCtx,
		createdService.ID)
	assert.Nil(err)
	assert.NotNil(pluginsFromKong)
	assert.Equal(2, len(pluginsFromKong))

	pluginsFromKong, err = client.Plugins.ListAllForRoute(defaultCtx,
		createdRoute.ID)
	assert.Nil(err)
	assert.NotNil(pluginsFromKong)
	assert.Equal(2, len(pluginsFromKong))

	for i := 0; i < len(plugins); i++ {
		assert.Nil(client.Plugins.Delete(defaultCtx, plugins[i].ID))
	}

	assert.Nil(client.Consumers.Delete(defaultCtx, createdConsumer.ID))
	assert.Nil(client.Routes.Delete(defaultCtx, createdRoute.ID))
	assert.Nil(client.Services.Delete(defaultCtx, createdService.ID))
}

func comparePlugins(expected, actual []*Plugin) bool {
	var expectedNames, actualNames []string
	for _, plugin := range expected {
		expectedNames = append(expectedNames, *plugin.Name)
	}

	for _, plugin := range actual {
		actualNames = append(actualNames, *plugin.Name)
	}

	return (compareSlices(expectedNames, actualNames))
}
