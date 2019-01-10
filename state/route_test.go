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
	assert.Panics(func() {
		collection.GetAll()
	})
}

// Regression test
// to ensure that the memory reference of the pointer returned by Get()
// is different from the one stored in MemDB.
func TestRouteGetMemoryReference(t *testing.T) {
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

	re, err = collection.Get("my-route")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Nil(re.SNIs)
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

func TestRouteGetAllByServiceName(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewRoutesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	targets := []*Route{
		{
			Route: kong.Route{
				ID:   kong.String("target1-id"),
				Name: kong.String("target1-name"),
				Service: &kong.Service{
					ID:   kong.String("upstream1-id"),
					Name: kong.String("upstream1-name"),
				},
			},
		},
		{
			Route: kong.Route{
				ID:   kong.String("target2-id"),
				Name: kong.String("target2-name"),
				Service: &kong.Service{
					ID:   kong.String("upstream1-id"),
					Name: kong.String("upstream1-name"),
				},
			},
		},
		{
			Route: kong.Route{
				ID:   kong.String("target3-id"),
				Name: kong.String("target3-name"),
				Service: &kong.Service{
					ID:   kong.String("upstream2-id"),
					Name: kong.String("upstream2-name"),
				},
			},
		},
		{
			Route: kong.Route{
				ID:   kong.String("target4-id"),
				Name: kong.String("target4-name"),
				Service: &kong.Service{
					ID:   kong.String("upstream2-id"),
					Name: kong.String("upstream2-name"),
				},
			},
		},
	}

	for _, target := range targets {
		err = collection.Add(*target)
		assert.Nil(err)
	}

	targets, err = collection.GetAllByServiceID("upstream1-id")
	assert.Nil(err)
	assert.Equal(2, len(targets))

	targets, err = collection.GetAllByServiceName("upstream2-name")
	assert.Nil(err)
	assert.Equal(2, len(targets))

	targets, err = collection.GetAllByServiceName("upstream1-id")
	assert.Nil(err)
	assert.Equal(0, len(targets))
}
