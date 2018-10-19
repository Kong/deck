package kong

import (
	"os"
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

func TestRoot(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	root, err := client.Root(nil)
	assert.Nil(err)
	assert.NotNil(root)
	assert.NotNil(root["version"])
}

func TestMain(m *testing.M) {
	// to test ListAll code for pagination
	pageSize = 1
	os.Exit(m.Run())
}
