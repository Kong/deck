package kong

import (
	"reflect"
	"sort"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestConsumersService(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	consumer := &Consumer{
		Username: String("foo"),
		CustomID: String("custom_id_foo"),
	}

	createdConsumer, err := client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(createdConsumer)

	consumer, err = client.Consumers.Get(defaultCtx, createdConsumer.ID)
	assert.Nil(err)
	assert.NotNil(consumer)

	consumer, err = client.Consumers.GetByCustomID(defaultCtx,
		String("does-not-exist"))
	assert.NotNil(err)
	assert.Nil(consumer)

	consumer, err = client.Consumers.GetByCustomID(defaultCtx,
		String("custom_id_foo"))
	assert.Nil(err)
	assert.NotNil(consumer)

	consumer.Username = String("bar")
	consumer, err = client.Consumers.Update(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)
	assert.Equal("bar", *consumer.Username)

	err = client.Consumers.Delete(defaultCtx, createdConsumer.ID)
	assert.Nil(err)

	// ID can be specified
	id := uuid.NewV4().String()
	consumer = &Consumer{
		Username: String("foo"),
		ID:       String(id),
	}

	createdConsumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(createdConsumer)
	assert.Equal(id, *createdConsumer.ID)

	err = client.Consumers.Delete(defaultCtx, createdConsumer.ID)
	assert.Nil(err)
}

func TestConsumerWithTags(T *testing.T) {
	runWhenKong(T, ">=1.1.0")
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	consumer := &Consumer{
		Username: String("foo"),
		Tags:     StringSlice("tag1", "tag2"),
	}

	createdConsumer, err := client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(createdConsumer)
	assert.Equal(StringSlice("tag1", "tag2"), createdConsumer.Tags)

	err = client.Consumers.Delete(defaultCtx, createdConsumer.ID)
	assert.Nil(err)
}

func TestConsumerListEndpoint(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// fixtures
	consumers := []*Consumer{
		{
			Username: String("foo1"),
		},
		{
			Username: String("foo2"),
		},
		{
			Username: String("foo3"),
		},
	}

	// create fixturs
	for i := 0; i < len(consumers); i++ {
		consumer, err := client.Consumers.Create(defaultCtx, consumers[i])
		assert.Nil(err)
		assert.NotNil(consumer)
		consumers[i] = consumer
	}

	consumersFromKong, next, err := client.Consumers.List(defaultCtx, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(consumersFromKong)
	assert.Equal(3, len(consumersFromKong))

	// check if we see all consumers
	assert.True(compareConsumers(consumers, consumersFromKong))

	// Test pagination
	consumersFromKong = []*Consumer{}

	// first page
	page1, next, err := client.Consumers.List(defaultCtx, &ListOpt{Size: 1})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))
	consumersFromKong = append(consumersFromKong, page1...)

	// last page
	next.Size = 2
	page2, next, err := client.Consumers.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(page2)
	assert.Equal(2, len(page2))
	consumersFromKong = append(consumersFromKong, page2...)

	assert.True(compareConsumers(consumers, consumersFromKong))

	consumers, err = client.Consumers.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(consumers)
	assert.Equal(3, len(consumers))

	for i := 0; i < len(consumers); i++ {
		assert.Nil(client.Consumers.Delete(defaultCtx, consumers[i].ID))
	}
}

func TestConsumerListWithTags(T *testing.T) {
	runWhenKong(T, ">=1.1.0")
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// fixtures
	consumers := []*Consumer{
		{
			Username: String("user1"),
			Tags:     StringSlice("tag1", "tag2"),
		},
		{
			Username: String("user2"),
			Tags:     StringSlice("tag2", "tag3"),
		},
		{
			Username: String("user3"),
			Tags:     StringSlice("tag1", "tag3"),
		},
		{
			Username: String("user4"),
			Tags:     StringSlice("tag1", "tag2"),
		},
		{
			Username: String("user5"),
			Tags:     StringSlice("tag2", "tag3"),
		},
		{
			Username: String("user6"),
			Tags:     StringSlice("tag1", "tag3"),
		},
	}

	// create fixtures
	for i := 0; i < len(consumers); i++ {
		consumer, err := client.Consumers.Create(defaultCtx, consumers[i])
		assert.Nil(err)
		assert.NotNil(consumer)
		consumers[i] = consumer
	}

	consumersFromKong, next, err := client.Consumers.List(defaultCtx, &ListOpt{
		Tags: StringSlice("tag1"),
	})
	assert.Nil(err)
	assert.Nil(next)
	assert.Equal(4, len(consumersFromKong))

	consumersFromKong, next, err = client.Consumers.List(defaultCtx, &ListOpt{
		Tags: StringSlice("tag2"),
	})
	assert.Nil(err)
	assert.Nil(next)
	assert.Equal(4, len(consumersFromKong))

	consumersFromKong, next, err = client.Consumers.List(defaultCtx, &ListOpt{
		Tags: StringSlice("tag1", "tag2"),
	})
	assert.Nil(err)
	assert.Nil(next)
	assert.Equal(6, len(consumersFromKong))

	consumersFromKong, next, err = client.Consumers.List(defaultCtx, &ListOpt{
		Tags:         StringSlice("tag1", "tag2"),
		MatchAllTags: true,
	})
	assert.Nil(err)
	assert.Nil(next)
	assert.Equal(2, len(consumersFromKong))

	consumersFromKong, next, err = client.Consumers.List(defaultCtx, &ListOpt{
		Tags: StringSlice("tag1", "tag2"),
		Size: 3,
	})
	assert.Nil(err)
	assert.NotNil(next)
	assert.Equal(3, len(consumersFromKong))

	consumersFromKong, next, err = client.Consumers.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.Equal(3, len(consumersFromKong))

	consumersFromKong, next, err = client.Consumers.List(defaultCtx, &ListOpt{
		Tags:         StringSlice("tag1", "tag2"),
		MatchAllTags: true,
		Size:         1,
	})
	assert.Nil(err)
	assert.NotNil(next)
	assert.Equal(1, len(consumersFromKong))

	consumersFromKong, next, err = client.Consumers.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.Equal(1, len(consumersFromKong))

	for i := 0; i < len(consumers); i++ {
		assert.Nil(client.Consumers.Delete(defaultCtx, consumers[i].Username))
	}
}

func compareConsumers(expected, actual []*Consumer) bool {
	var expectedUsernames, actualUsernames []string
	for _, consumer := range expected {
		expectedUsernames = append(expectedUsernames, *consumer.Username)
	}

	for _, consumer := range actual {
		actualUsernames = append(actualUsernames, *consumer.Username)
	}

	return (compareSlices(expectedUsernames, actualUsernames))
}

func compareSlices(expected, actual []string) bool {
	sort.Strings(expected)
	sort.Strings(actual)
	return (reflect.DeepEqual(expected, actual))
}
