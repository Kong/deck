package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsumersService(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	consumer := &Consumer{
		Username: String("foo"),
	}

	createdConsumer, err := client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(createdConsumer)

	consumer, err = client.Consumers.Get(defaultCtx, createdConsumer.ID)
	assert.Nil(err)
	assert.NotNil(consumer)

	consumer.Username = String("bar")
	consumer, err = client.Consumers.Update(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)
	assert.Equal("bar", *consumer.Username)

	err = client.Consumers.Delete(defaultCtx, createdConsumer.ID)
	assert.Nil(err)
}
