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
