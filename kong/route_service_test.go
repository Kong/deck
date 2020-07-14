package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestRoutesRoute(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	route := &Route{}

	routeNotCreated, err := client.Routes.Create(defaultCtx, route)
	assert.NotNil(err)
	assert.Nil(routeNotCreated)

	// service for the route
	service := &Service{
		Name: String("foo2"),
		Host: String("upstream"),
		Port: Int(42),
		Path: String("/path"),
	}

	service, err = client.Services.Create(defaultCtx, service)
	assert.Nil(err)
	assert.NotNil(service)

	route = &Route{
		Hosts:   StringSlice("host1.com", "host2.com"),
		Service: service,
	}
	createdRoute, err := client.Routes.Create(defaultCtx, route)
	assert.Nil(err)
	assert.NotNil(createdRoute)

	route, err = client.Routes.Get(defaultCtx, createdRoute.ID)
	assert.Nil(err)
	assert.NotNil(route)
	assert.Empty(route.Methods)
	assert.Empty(route.Paths)

	route.Hosts = StringSlice("newHost.com")
	route.Methods = StringSlice("GET", "POST")
	route, err = client.Routes.Update(defaultCtx, route)
	assert.Nil(err)
	assert.NotNil(route)
	assert.Equal(1, len(route.Hosts))
	assert.Equal("newHost.com", *route.Hosts[0])

	err = client.Routes.Delete(defaultCtx, createdRoute.ID)
	assert.Nil(err)

	// ID can be specified
	id := uuid.NewV4().String()
	route = &Route{
		ID:        String(id),
		Name:      String("new-route"),
		SNIs:      StringSlice("snihost1.com", "snihost2.com"),
		Protocols: StringSlice("tcp", "tls"),
		Destinations: []*CIDRPort{
			{
				IP:   String("10.0.0.0/8"),
				Port: Int(80),
			},
		},
		Service: service,
	}

	createdRoute, err = client.Routes.Create(defaultCtx, route)
	assert.Nil(err)
	assert.NotNil(createdRoute)
	assert.Equal(id, *createdRoute.ID)
	assert.Equal(2, len(createdRoute.SNIs))
	assert.Equal("snihost1.com", *createdRoute.SNIs[0])
	assert.Equal("snihost2.com", *createdRoute.SNIs[1])
	assert.Equal("10.0.0.0/8", *createdRoute.Destinations[0].IP)
	assert.Equal(80, *createdRoute.Destinations[0].Port)

	err = client.Routes.Delete(defaultCtx, createdRoute.ID)
	assert.Nil(err)

	err = client.Services.Delete(defaultCtx, service.ID)
	assert.Nil(err)

	_, err = client.Routes.Create(defaultCtx, nil)
	assert.NotNil(err)

	_, err = client.Routes.Update(defaultCtx, nil)
	assert.NotNil(err)
}

func TestRouteWithTags(T *testing.T) {
	runWhenKong(T, ">=1.1.0")
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	route := &Route{
		Name:  String("key-auth"),
		Paths: StringSlice("/"),
		Tags:  StringSlice("tag1", "tag2"),
	}

	createdRoute, err := client.Routes.Create(defaultCtx, route)
	assert.Nil(err)
	assert.NotNil(createdRoute)
	assert.Equal(StringSlice("tag1", "tag2"), createdRoute.Tags)

	err = client.Routes.Delete(defaultCtx, createdRoute.ID)
	assert.Nil(err)
}

func TestCreateInRoute(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	service := &Service{
		Name: String("foo"),
		Host: String("upstream"),
		Port: Int(42),
		Path: String("/path"),
	}

	createdService, err := client.Services.Create(defaultCtx, service)
	assert.Nil(err)
	assert.NotNil(createdService)

	route := &Route{
		Hosts: StringSlice("host1.com", "host2.com"),
	}

	// specifying name won't work
	routeNotCreated, err := client.Routes.CreateInService(defaultCtx,
		createdService.Name, route)
	assert.Nil(routeNotCreated)
	assert.NotNil(err)

	createdRoute, err := client.Routes.CreateInService(defaultCtx,
		createdService.ID, route)
	assert.Nil(err)
	assert.NotNil(createdRoute)

	assert.Nil(client.Routes.Delete(defaultCtx, createdRoute.ID))
	assert.Nil(client.Services.Delete(defaultCtx, createdService.ID))
}
func TestRouteListEndpoint(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	service := &Service{
		Name: String("foo"),
		Host: String("upstream"),
		Port: Int(42),
		Path: String("/path"),
	}

	createdService, err := client.Services.Create(defaultCtx, service)
	assert.Nil(err)
	assert.NotNil(createdService)

	// fixtures
	routes := []*Route{
		{
			Paths:   StringSlice("/foo1"),
			Service: createdService,
		},
		{
			Paths:   StringSlice("/foo2"),
			Service: createdService,
		},
		{
			Paths:   StringSlice("/foo3"),
			Service: createdService,
		},
	}

	// create fixturs
	for i := 0; i < len(routes); i++ {
		route, err := client.Routes.Create(defaultCtx, routes[i])
		assert.Nil(err)
		assert.NotNil(route)
		routes[i] = route
	}

	routesFromKong, next, err := client.Routes.List(defaultCtx, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(routesFromKong)
	assert.Equal(3, len(routesFromKong))

	// check if we see all routes
	assert.True(compareRoutes(routes, routesFromKong))

	// Test pagination
	routesFromKong = []*Route{}

	// first page
	page1, next, err := client.Routes.List(defaultCtx, &ListOpt{Size: 1})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))
	routesFromKong = append(routesFromKong, page1...)

	// last page
	next.Size = 2
	page2, next, err := client.Routes.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(page2)
	assert.Equal(2, len(page2))
	routesFromKong = append(routesFromKong, page2...)

	assert.True(compareRoutes(routes, routesFromKong))

	routesForService, next, err := client.Routes.ListForService(defaultCtx,
		createdService.ID, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(routesForService)
	assert.True(compareRoutes(routes, routesForService))

	routes, err = client.Routes.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(routes)
	assert.Equal(3, len(routes))

	for i := 0; i < len(routes); i++ {
		assert.Nil(client.Routes.Delete(defaultCtx, routes[i].ID))
	}

	assert.Nil(client.Services.Delete(defaultCtx, createdService.ID))
}

func compareRoutes(expected, actual []*Route) bool {
	var expectedUsernames, actualUsernames []string
	for _, route := range expected {
		expectedUsernames = append(expectedUsernames, *route.Paths[0])
	}

	for _, route := range actual {
		actualUsernames = append(actualUsernames, *route.Paths[0])
	}

	return (compareSlices(expectedUsernames, actualUsernames))
}

func TestRouteWithHeaders(T *testing.T) {
	runWhenKong(T, ">=1.3.0")
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	route := &Route{
		Name: String("route-by-header"),
		Headers: map[string][]string{
			"foo": {"bar"},
		},
		Tags: StringSlice("tag1", "tag2"),
	}

	createdRoute, err := client.Routes.Create(defaultCtx, route)
	assert.Nil(err)
	assert.NotNil(createdRoute)
	assert.Equal(StringSlice("tag1", "tag2"), createdRoute.Tags)
	assert.Equal(map[string][]string{"foo": {"bar"}}, createdRoute.Headers)

	err = client.Routes.Delete(defaultCtx, createdRoute.ID)
	assert.Nil(err)
}
