package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPISeviceCreate(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	api := &API{
		URIs:        []*string{String("/test")},
		Name:        String("test"),
		UpstreamURL: String("https://google.com"),
	}

	createdAPI, err := client.APIs.Create(defaultCtx, api)
	assert.Nil(err)
	assert.NotNil(createdAPI)

	api, err = client.APIs.Get(defaultCtx, createdAPI.ID)
	assert.Nil(err)
	assert.NotNil(api)

	api.Methods = []*string{String("GET")}
	api, err = client.APIs.Update(defaultCtx, api)
	assert.Nil(err)
	assert.NotNil(api)
	assert.Equal(1, len(api.Methods))
	assert.Equal("GET", *api.Methods[0])

	err = client.APIs.Delete(defaultCtx, createdAPI.ID)
	assert.Nil(err)
}

func TestAPIListEndpoint(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// fixtures
	apis := []*API{
		{
			Name:        String("foo1"),
			UpstreamURL: String("http://upstream:80/foo1"),
			URIs:        StringSlice("/foo1"),
		},
		{
			Name:        String("foo2"),
			UpstreamURL: String("http://upstream:80/foo2"),
			URIs:        StringSlice("/foo2"),
		},
		{
			Name:        String("foo3"),
			UpstreamURL: String("http://upstream:80/foo3"),
			URIs:        StringSlice("/foo3"),
		},
	}

	// create fixturs
	for i := 0; i < len(apis); i++ {
		api, err := client.APIs.Create(defaultCtx, apis[i])
		assert.Nil(err)
		assert.NotNil(api)
		apis[i] = api
	}

	apisFromKong, next, err := client.APIs.List(defaultCtx, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(apisFromKong)
	assert.Equal(3, len(apisFromKong))

	// check if we see all apis
	assert.True(compareAPIs(apis, apisFromKong))

	// Test pagination
	apisFromKong = []*API{}

	// first page
	page1, next, err := client.APIs.List(defaultCtx, &ListOpt{Size: 1})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))
	apisFromKong = append(apisFromKong, page1...)

	// intermediate page
	// Old DAO can't do dynamic paging it seems
	page2, next, err := client.APIs.List(defaultCtx, next)
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page2)
	assert.Equal(1, len(page2))
	apisFromKong = append(apisFromKong, page2...)

	// last page
	page3, next, err := client.APIs.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(page3)
	assert.Equal(1, len(page2))
	apisFromKong = append(apisFromKong, page3...)

	assert.True(compareAPIs(apis, apisFromKong))

	apis, err = client.APIs.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(apis)
	assert.Equal(3, len(apis))

	for i := 0; i < len(apis); i++ {
		assert.Nil(client.APIs.Delete(defaultCtx, apis[i].ID))
	}
}

func compareAPIs(expected, actual []*API) bool {
	var expectedNames, actualNames []string
	for _, api := range expected {
		expectedNames = append(expectedNames, *api.Name)
	}

	for _, api := range actual {
		actualNames = append(actualNames, *api.Name)
	}

	return (compareSlices(expectedNames, actualNames))
}
