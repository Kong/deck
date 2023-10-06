package state

import (
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func consumersCollection() *ConsumersCollection {
	return state().Consumers
}

func TestConsumerInsert(t *testing.T) {
	assert := assert.New(t)
	collection := consumersCollection()

	var consumer Consumer

	assert.NotNil(collection.Add(consumer))

	consumer.ID = kong.String("first")
	assert.Nil(collection.Add(consumer))

	// re-insert
	consumer.Username = kong.String("my-name")
	assert.NotNil(collection.Add(consumer))
}

func TestConsumerGetUpdate(t *testing.T) {
	assert := assert.New(t)
	collection := consumersCollection()

	var consumer Consumer
	consumer.ID = kong.String("first")
	consumer.Username = kong.String("my-name")
	err := collection.Add(consumer)
	assert.Nil(err)

	c, err := collection.GetByIDOrUsername("")
	assert.NotNil(err)
	assert.Nil(c)

	c, err = collection.GetByIDOrUsername("first")
	assert.Nil(err)
	assert.NotNil(c)

	c.ID = nil
	c.Username = kong.String("my-updated-name")
	err = collection.Update(*c)
	assert.NotNil(err)

	c.ID = kong.String("does-not-exist")
	assert.NotNil(collection.Update(*c))

	c.ID = kong.String("first")
	assert.Nil(collection.Update(*c))

	c, err = collection.GetByIDOrUsername("my-name")
	assert.NotNil(err)
	assert.Nil(c)

	c, err = collection.GetByIDOrUsername("my-updated-name")
	assert.Nil(err)
	assert.NotNil(c)
}

// Test to ensure that the memory reference of the pointer returned by Get()
// is different from the one stored in MemDB.
func TestConsumerGetMemoryReference(t *testing.T) {
	assert := assert.New(t)
	collection := consumersCollection()

	var consumer Consumer
	consumer.ID = kong.String("first")
	consumer.Username = kong.String("my-name")
	err := collection.Add(consumer)
	assert.Nil(err)

	c, err := collection.GetByIDOrUsername("first")
	assert.Nil(err)
	assert.NotNil(c)
	c.Username = kong.String("update-should-not-reflect")

	c, err = collection.GetByIDOrUsername("first")
	assert.Nil(err)
	assert.Equal("my-name", *c.Username)
}

func TestConsumersInvalidType(t *testing.T) {
	assert := assert.New(t)
	collection := consumersCollection()

	type c2 Consumer
	var c c2
	c.Username = kong.String("my-name")
	c.ID = kong.String("first")
	txn := collection.db.Txn(true)
	assert.Nil(txn.Insert(consumerTableName, &c))
	txn.Commit()

	assert.Panics(func() {
		collection.GetByIDOrUsername("my-name")
	})
	assert.Panics(func() {
		collection.GetAll()
	})
}

func TestConsumerDelete(t *testing.T) {
	assert := assert.New(t)
	collection := consumersCollection()

	var consumer Consumer
	consumer.ID = kong.String("first")
	consumer.Username = kong.String("my-consumer")
	err := collection.Add(consumer)
	assert.Nil(err)

	c, err := collection.GetByIDOrUsername("my-consumer")
	assert.Nil(err)
	assert.NotNil(c)
	assert.Equal("first", *c.ID)

	err = collection.Delete("first")
	assert.Nil(err)

	err = collection.Delete("")
	assert.NotNil(err)

	err = collection.Delete(*c.ID)
	assert.NotNil(err)
}

func TestConsumerGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := consumersCollection()

	consumers := []Consumer{
		{
			Consumer: kong.Consumer{
				ID:       kong.String("first"),
				Username: kong.String("my-consumer1"),
			},
		},
		{
			Consumer: kong.Consumer{
				ID:       kong.String("second"),
				Username: kong.String("my-consumer2"),
			},
		},
	}
	for _, s := range consumers {
		assert.Nil(collection.Add(s))
	}

	allConsumers, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(len(consumers), len(allConsumers))
}
