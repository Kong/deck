package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestRoutesRoute(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	route := &Route{
		Hosts: StringSlice("host1.com", "host2.com"),
	}

	// foreign key not specified
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
		ID:      String(id),
		Hosts:   StringSlice("buzz"),
		Service: service,
	}

	createdRoute, err = client.Routes.Create(defaultCtx, route)
	assert.Nil(err)
	assert.NotNil(createdRoute)
	assert.Equal(id, *createdRoute.ID)
	assert.Equal(1, len(createdRoute.Hosts))
	assert.Equal("buzz", *createdRoute.Hosts[0])

	err = client.Routes.Delete(defaultCtx, createdRoute.ID)
	assert.Nil(err)

	err = client.Services.Delete(defaultCtx, service.ID)
	assert.Nil(err)

	_, err = client.Routes.Create(defaultCtx, nil)
	assert.NotNil(err)

	_, err = client.Routes.Update(defaultCtx, nil)
	assert.NotNil(err)
}

func TestCreateInRoute(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
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
	routeNotCreated, err := client.Routes.CreateInService(defaultCtx, createdService.Name, route)
	assert.Nil(routeNotCreated)
	assert.NotNil(err)

	createdRoute, err := client.Routes.CreateInService(defaultCtx, createdService.ID, route)
	assert.Nil(err)
	assert.NotNil(createdRoute)

	assert.Nil(client.Routes.Delete(defaultCtx, createdRoute.ID))
	assert.Nil(client.Services.Delete(defaultCtx, createdService.ID))
}
func TestRouteListEndpoint(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
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

	routesForService, next, err := client.Routes.ListForService(defaultCtx, createdService.ID, nil)
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
