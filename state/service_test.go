package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func TestServiceInsert(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewServicesCollection()
	assert.Nil(err)
	assert.NotNil(collection)
	var service Service
	service.ID = kong.String("first")
	service.Host = kong.String("example.com")
	err = collection.Add(service)
	assert.NotNil(err)

	var service2 Service
	service2.Name = kong.String("my-service")
	service2.ID = kong.String("first")
	service2.Host = kong.String("example.com")
	assert.NotNil(service2.Service)
	err = collection.Add(service2)
	assert.NotNil(service2.Service)
	assert.Nil(err)
}

func TestServiceGetUpdate(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewServicesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var service Service
	service.Name = kong.String("my-service")
	service.ID = kong.String("first")
	err = collection.Add(service)
	assert.Nil(err)

	se, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(se)
	se.Host = kong.String("example.com")
	err = collection.Update(*se)
	assert.Nil(err)

	se, err = collection.Get("my-service")
	assert.Nil(err)
	assert.NotNil(se)
	assert.Equal("example.com", *se.Host)
}

// Regression test
// to ensure that the memory reference of the pointer returned by Get()
// is different from the one stored in MemDB.
func TestServiceGetMemoryReference(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewServicesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var service Service
	service.Name = kong.String("my-service")
	service.ID = kong.String("first")
	err = collection.Add(service)
	assert.Nil(err)

	se, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(se)
	se.Host = kong.String("example.com")

	se, err = collection.Get("my-service")
	assert.Nil(err)
	assert.NotNil(se)
	assert.Nil(se.Host)
}

func TestServicesInvalidType(t *testing.T) {
	assert := assert.New(t)

	collection, err := NewServicesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var route Route
	route.Name = kong.String("my-route")
	route.ID = kong.String("first")
	txn := collection.memdb.Txn(true)
	txn.Insert(serviceTableName, &route)
	txn.Commit()

	assert.Panics(func() {
		collection.Get("my-route")
	})
	assert.Panics(func() {
		collection.GetAll()
	})
}

func TestServiceDelete(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewServicesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var service Service
	service.Name = kong.String("my-service")
	service.ID = kong.String("first")
	service.Host = kong.String("example.com")
	err = collection.Add(service)
	assert.Nil(err)

	se, err := collection.Get("my-service")
	assert.Nil(err)
	assert.NotNil(se)
	assert.Equal("example.com", *se.Host)

	err = collection.Delete(*se.ID)
	assert.Nil(err)

	err = collection.Delete(*se.ID)
	assert.NotNil(err)
}

func TestServiceGetAll(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewServicesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	services := []Service{
		{
			Service: kong.Service{
				ID:   kong.String("first"),
				Name: kong.String("my-service1"),
				Host: kong.String("example.com"),
			},
		},
		{
			Service: kong.Service{
				ID:   kong.String("second"),
				Name: kong.String("my-service2"),
				Host: kong.String("example.com"),
			},
		},
	}
	for _, s := range services {
		assert.Nil(collection.Add(s))
	}

	allServices, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(len(services), len(allServices))
}

// Regression test
// to ensure that the memory reference of the pointer returned by Get()
// is different from the one stored in MemDB.
func TestServiceGetAllMemoryReference(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewServicesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	services := []Service{
		{
			Service: kong.Service{
				ID:   kong.String("first"),
				Name: kong.String("my-service1"),
				Host: kong.String("example.com"),
			},
		},
		{
			Service: kong.Service{
				ID:   kong.String("second"),
				Name: kong.String("my-service2"),
				Host: kong.String("example.com"),
			},
		},
	}
	for _, s := range services {
		assert.Nil(collection.Add(s))
	}

	allServices, err := collection.GetAll()
	assert.Nil(err)
	assert.Equal(len(services), len(allServices))

	allServices[0].Host = kong.String("new.example.com")
	allServices[1].Host = kong.String("new.example.com")

	service, err := collection.Get("my-service1")
	assert.Nil(err)
	assert.NotNil(service)
	assert.Equal("example.com", *service.Host)

}
