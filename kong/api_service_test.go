package kong

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPISeviceCreate(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	var uris []*string
	uri := string("/test")
	uris = append(uris, &uri)
	name := "test"
	url := "https://google.com"
	api := &API{
		URIs:        uris,
		Name:        &name,
		UpstreamURL: &url,
	}

	createdAPI, err := client.APIService.Create(context.Background(), api)
	assert.Nil(err)
	assert.NotNil(createdAPI)

}
