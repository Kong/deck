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

	api := &API{
		URIs:        []*string{String("/test")},
		Name:        String("test"),
		UpstreamURL: String("https://google.com"),
	}

	createdAPI, err := client.APIService.Create(context.Background(), api)
	assert.Nil(err)
	assert.NotNil(createdAPI)

}
