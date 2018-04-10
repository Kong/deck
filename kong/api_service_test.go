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
