package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func basicAuthsCollection() *BasicAuthsCollection {
	return state().BasicAuths
}

func TestBasicAuthInsert(t *testing.T) {
	assert := assert.New(t)
	collection := basicAuthsCollection()

	var basicAuth BasicAuth
	basicAuth.ID = kong.String("first")
	err := collection.Add(basicAuth)
	assert.NotNil(err)

	basicAuth.Username = kong.String("my-username")
	err = collection.Add(basicAuth)
	assert.NotNil(err)

	var basicAuth2 BasicAuth
	basicAuth2.Username = kong.String("my-username")
	basicAuth2.ID = kong.String("first")
	basicAuth2.Consumer = &kong.Consumer{
		ID:       kong.String("consumer-id"),
		Username: kong.String("my-username"),
	}
	err = collection.Add(basicAuth2)
	assert.Nil(err)
}

func TestBasicAuthGet(t *testing.T) {
	assert := assert.New(t)
	collection := basicAuthsCollection()

	var basicAuth BasicAuth
	basicAuth.Username = kong.String("my-username")
	basicAuth.ID = kong.String("first")
	basicAuth.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}

	err := collection.Add(basicAuth)
	assert.Nil(err)

	res, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("my-username", *res.Username)

	res, err = collection.Get("my-username")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("first", *res.ID)
	assert.Equal("consumer1-id", *res.Consumer.ID)

	res, err = collection.Get("does-not-exist")
	assert.NotNil(err)
	assert.Nil(res)
}

func TestBasicAuthUpdate(t *testing.T) {
	assert := assert.New(t)
	collection := basicAuthsCollection()

	var basicAuth BasicAuth
	basicAuth.Username = kong.String("my-username")
	basicAuth.ID = kong.String("first")
	basicAuth.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}

	err := collection.Add(basicAuth)
	assert.Nil(err)

	res, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("my-username", *res.Username)

	res.Username = kong.String("my-username2")
	res.Password = kong.String("password")
	err = collection.Update(*res)
	assert.Nil(err)

	res, err = collection.Get("my-username")
	assert.NotNil(err)

	res, err = collection.Get("my-username2")
	assert.Nil(err)
	assert.Equal("first", *res.ID)
	assert.Equal("password", *res.Password)
}

func TestBasicAuthDelete(t *testing.T) {
	assert := assert.New(t)
	collection := basicAuthsCollection()

	var basicAuth BasicAuth
	basicAuth.Username = kong.String("my-username1")
	basicAuth.ID = kong.String("first")
	basicAuth.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}
	err := collection.Add(basicAuth)
	assert.Nil(err)

	res, err := collection.Get("my-username1")
	assert.Nil(err)
	assert.NotNil(res)

	err = collection.Delete(*res.ID)
	assert.Nil(err)

	res, err = collection.Get("my-username1")
	assert.NotNil(err)
	assert.Nil(res)

	// delete a non-existing one
	err = collection.Delete("first")
	assert.NotNil(err)

	err = collection.Delete("my-username1")
	assert.NotNil(err)
}

func TestBasicAuthGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := basicAuthsCollection()

	populateWithBasicAuthFixtures(assert, collection)

	basicAuths, err := collection.GetAll()
	assert.Nil(err)
	assert.Equal(5, len(basicAuths))
}

func TestBasicAuthGetByConsumer(t *testing.T) {
	assert := assert.New(t)
	collection := basicAuthsCollection()

	populateWithBasicAuthFixtures(assert, collection)

	basicAuths, err := collection.GetAllByConsumerID("consumer1-id")
	assert.Nil(err)
	assert.Equal(3, len(basicAuths))

	basicAuths, err = collection.GetAllByConsumerUsername("consumer2-name")
	assert.Nil(err)
	assert.Equal(2, len(basicAuths))
}

func populateWithBasicAuthFixtures(assert *assert.Assertions,
	collection *BasicAuthsCollection) {

	basicAuths := []BasicAuth{
		{
			BasicAuth: kong.BasicAuth{
				Username: kong.String("my-username11"),
				ID:       kong.String("first"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			BasicAuth: kong.BasicAuth{
				Username: kong.String("my-username12"),
				ID:       kong.String("second"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			BasicAuth: kong.BasicAuth{
				Username: kong.String("my-username13"),
				ID:       kong.String("third"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			BasicAuth: kong.BasicAuth{
				Username: kong.String("my-username21"),
				ID:       kong.String("fourth"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer2-id"),
					Username: kong.String("consumer2-name"),
				},
			},
		},
		{
			BasicAuth: kong.BasicAuth{
				Username: kong.String("my-username22"),
				ID:       kong.String("fifth"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer2-id"),
					Username: kong.String("consumer2-name"),
				},
			},
		},
	}

	for _, k := range basicAuths {
		err := collection.Add(k)
		assert.Nil(err)
	}
}
