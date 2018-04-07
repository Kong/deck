package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKongStatus(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	status, err := client.Status(nil)
	assert.Nil(err)
	assert.NotNil(status)
}
