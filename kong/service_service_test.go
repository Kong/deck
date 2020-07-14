package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestServicesService(T *testing.T) {
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

	service, err = client.Services.Get(defaultCtx, createdService.ID)
	assert.Nil(err)
	assert.NotNil(service)

	service.Name = String("bar")
	service.Host = String("newUpstream")
	service, err = client.Services.Update(defaultCtx, service)
	assert.Nil(err)
	assert.NotNil(service)
	assert.Equal("bar", *service.Name)
	assert.Equal("newUpstream", *service.Host)
	assert.Equal(42, *service.Port)

	route, err := client.Routes.CreateInService(defaultCtx, service.ID, &Route{
		Paths: StringSlice("/route"),
	})
	assert.Nil(err)
	assert.NotNil(route)

	serviceForRoute, err := client.Services.GetForRoute(defaultCtx, route.ID)
	assert.Nil(err)
	assert.NotNil(serviceForRoute)

	assert.Equal(*service.ID, *serviceForRoute.ID)

	err = client.Routes.Delete(defaultCtx, route.ID)
	assert.Nil(err)

	err = client.Services.Delete(defaultCtx, service.ID)
	assert.Nil(err)

	// ID can be specified
	id := uuid.NewV4().String()
	service = &Service{
		Name: String("fizz"),
		ID:   String(id),
		Host: String("buzz"),
	}

	createdService, err = client.Services.Create(defaultCtx, service)
	assert.Nil(err)
	assert.NotNil(createdService)
	assert.Equal(id, *createdService.ID)
	assert.Equal("buzz", *createdService.Host)

	err = client.Services.Delete(defaultCtx, createdService.ID)
	assert.Nil(err)

	_, err = client.Services.Create(defaultCtx, nil)
	assert.NotNil(err)

	_, err = client.Services.Update(defaultCtx, nil)
	assert.NotNil(err)
}

func TestServiceWithTags(T *testing.T) {
	runWhenKong(T, ">=1.1.0")
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	service := &Service{
		Name: String("key-auth"),
		Host: String("example.com"),
		Tags: StringSlice("tag1", "tag2"),
	}

	createdService, err := client.Services.Create(defaultCtx, service)
	assert.Nil(err)
	assert.NotNil(createdService)
	assert.Equal(StringSlice("tag1", "tag2"), createdService.Tags)

	err = client.Services.Delete(defaultCtx, createdService.ID)
	assert.Nil(err)
}

func TestServiceListEndpoint(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// fixtures
	services := []*Service{
		{
			Name: String("foo1"),
			Host: String("upstream1.com"),
		},
		{
			Name: String("foo2"),
			Host: String("upstream2.com"),
		},
		{
			Name: String("foo3"),
			Host: String("upstream3.com"),
		},
	}

	// create fixturs
	for i := 0; i < len(services); i++ {
		service, err := client.Services.Create(defaultCtx, services[i])
		assert.Nil(err)
		assert.NotNil(service)
		services[i] = service
	}

	servicesFromKong, next, err := client.Services.List(defaultCtx, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(servicesFromKong)
	assert.Equal(3, len(servicesFromKong))

	// check if we see all services
	assert.True(compareServices(services, servicesFromKong))

	// Test pagination
	servicesFromKong = []*Service{}

	// first page
	page1, next, err := client.Services.List(defaultCtx, &ListOpt{Size: 1})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))
	servicesFromKong = append(servicesFromKong, page1...)

	// last page
	next.Size = 2
	page2, next, err := client.Services.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(page2)
	assert.Equal(2, len(page2))
	servicesFromKong = append(servicesFromKong, page2...)

	assert.True(compareServices(services, servicesFromKong))

	services, err = client.Services.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(services)
	assert.Equal(3, len(services))

	for i := 0; i < len(services); i++ {
		assert.Nil(client.Services.Delete(defaultCtx, services[i].ID))
	}
}

func compareServices(expected, actual []*Service) bool {
	var expectedUsernames, actualUsernames []string
	for _, service := range expected {
		expectedUsernames = append(expectedUsernames, *service.Name)
	}

	for _, service := range actual {
		actualUsernames = append(actualUsernames, *service.Name)
	}

	return (compareSlices(expectedUsernames, actualUsernames))
}

func TestServiceWithClientCert(T *testing.T) {
	runWhenKong(T, ">=1.3.0")
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	certificate := &Certificate{
		Key:  String(key1),
		Cert: String(cert1),
	}
	createdCertificate, err := client.Certificates.Create(defaultCtx, certificate)
	assert.Nil(err)
	assert.NotNil(createdCertificate)

	service := &Service{
		Name:              String("foo"),
		Host:              String("example.com"),
		Protocol:          String("https"),
		ClientCertificate: createdCertificate,
	}

	createdService, err := client.Services.Create(defaultCtx, service)
	assert.Nil(err)
	assert.NotNil(createdService)
	assert.Equal(*createdCertificate.ID, *createdService.ClientCertificate.ID)

	err = client.Services.Delete(defaultCtx, createdService.ID)
	assert.Nil(err)

	err = client.Certificates.Delete(defaultCtx, createdCertificate.ID)
	assert.Nil(err)
}
