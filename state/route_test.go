package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func TestBasicRoute(t *testing.T) {
	assert := assert.New(t)
	state, err := NewRoutesCollection()
	assert.Nil(err)
	assert.NotNil(state)

	var route Route
	// route.Name = kong.String("foo")
	route.ID = kong.String("first")
	route.Name = kong.String("my-route")
	route.Service = &kong.Service{
		Name: kong.String("prod"),
	}
	err = state.Add(route)
	assert.Nil(err)

	r, err := state.Get("first")
	assert.Nil(err)
	assert.NotNil(r)
	assert.Equal("prod", *r.Service.Name)

	r, err = state.Get("my-route")
	assert.Nil(err)
	assert.NotNil(r)
	assert.Equal("prod", *r.Service.Name)

	routes, err := state.GetAllRoutesByServiceName("prod")
	assert.Nil(err)
	assert.NotNil(routes)
	assert.Equal(1, len(routes))

	r = routes[0]
	assert.Equal("prod", *r.Service.Name)
	assert.Equal("first", *r.ID)
}
