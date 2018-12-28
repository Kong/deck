package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func TestRouteInsert(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewRoutesCollection()
	assert.Nil(err)
	assert.NotNil(collection)
	var route Route
	route.Name = kong.String("my-route")
	route.ID = kong.String("first")
	route.Hosts = kong.StringSlice("example.com", "demo.example.com")
	err = collection.Add(route)
	assert.NotNil(err)

	var route2 Route
	route2.Name = kong.String("my-route")
	route2.ID = kong.String("first")
	route2.Hosts = kong.StringSlice("example.com", "demo.example.com")
	route2.Service = &kong.Service{
		ID:   kong.String("service1-id"),
		Name: kong.String("service1-name"),
	}
	assert.NotNil(route2.Service)
	err = collection.Add(route2)
	assert.NotNil(route2.Service)
	assert.Nil(err)
}

func TestRouteGetUpdate(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewRoutesCollection()
	assert.Nil(err)
	assert.NotNil(collection)
	var route Route
	route.Name = kong.String("my-route")
	route.ID = kong.String("first")
	route.Hosts = kong.StringSlice("example.com", "demo.example.com")
	route.Service = &kong.Service{
		ID:   kong.String("service1-id"),
		Name: kong.String("service1-name"),
	}
	assert.NotNil(route.Service)
	err = collection.Add(route)
	assert.NotNil(route.Service)
	assert.Nil(err)

	re, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Equal("my-route", *re.Name)
	re.SNIs = kong.StringSlice("example.com", "demo.example.com")
	err = collection.Update(*re)
	assert.Nil(err)

	re, err = collection.Get("my-route")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Equal("example.com", *re.SNIs[0])
}

func TestRoutesInvalidType(t *testing.T) {
	assert := assert.New(t)

	collection, err := NewRoutesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var service Service
	service.Name = kong.String("my-service")
	service.ID = kong.String("first")
	txn := collection.memdb.Txn(true)
	txn.Insert(routeTableName, &service)
	txn.Commit()

	assert.Panics(func() {
		collection.Get("my-service")
	})
}

func TestRouteDelete(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewRoutesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var route Route
	route.Name = kong.String("my-route")
	route.ID = kong.String("first")
	route.Hosts = kong.StringSlice("example.com", "demo.example.com")
	route.Service = &kong.Service{
		ID:   kong.String("service1-id"),
		Name: kong.String("service1-name"),
	}
	err = collection.Add(route)
	assert.Nil(err)

	re, err := collection.Get("my-route")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Equal("example.com", *re.Hosts[0])

	err = collection.Delete(*re.ID)
	assert.Nil(err)

	err = collection.Delete(*re.ID)
	assert.NotNil(err)
}

func TestRouteGetAll(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewRoutesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var route Route
	route.Name = kong.String("my-route1")
	route.ID = kong.String("first")
	route.Hosts = kong.StringSlice("example.com", "demo.example.com")
	route.Service = &kong.Service{
		ID:   kong.String("service1-id"),
		Name: kong.String("service1-name"),
	}
	err = collection.Add(route)
	assert.Nil(err)

	var route2 Route
	route2.Name = kong.String("my-route2")
	route2.ID = kong.String("second")
	route2.Hosts = kong.StringSlice("example.com", "demo.example.com")
	route2.Service = &kong.Service{
		ID:   kong.String("service1-id"),
		Name: kong.String("service1-name"),
	}
	err = collection.Add(route2)
	assert.Nil(err)

	routes, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(2, len(routes))
}
