package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func keyAuthsCollection() *KeyAuthsCollection {
	return state().KeyAuths
}

func TestKeyAuthInsert(t *testing.T) {
	assert := assert.New(t)
	collection := keyAuthsCollection()

	var keyAuth KeyAuth
	keyAuth.Key = kong.String("my-secret-apikey")
	keyAuth.ID = kong.String("first")
	err := collection.Add(keyAuth)
	assert.NotNil(err)

	var keyAuth2 KeyAuth
	keyAuth2.Key = kong.String("my-secret-apikey")
	keyAuth2.ID = kong.String("first")
	keyAuth2.Consumer = &kong.Consumer{
		ID:       kong.String("consumer-id"),
		Username: kong.String("my-username"),
	}
	err = collection.Add(keyAuth2)
	assert.Nil(err)
}

func TestKeyAuthGet(t *testing.T) {
	assert := assert.New(t)
	collection := keyAuthsCollection()

	var keyAuth KeyAuth
	keyAuth.Key = kong.String("my-apikey")
	keyAuth.ID = kong.String("first")
	keyAuth.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}

	err := collection.Add(keyAuth)
	assert.Nil(err)

	res, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("my-apikey", *res.Key)

	res, err = collection.Get("my-apikey")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("first", *res.ID)
	assert.Equal("consumer1-id", *res.Consumer.ID)

	res, err = collection.Get("does-not-exist")
	assert.NotNil(err)
	assert.Nil(res)
}

func TestKeyAuthUpdate(t *testing.T) {
	assert := assert.New(t)
	collection := keyAuthsCollection()

	var keyAuth KeyAuth
	keyAuth.Key = kong.String("my-apikey")
	keyAuth.ID = kong.String("first")
	keyAuth.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}

	err := collection.Add(keyAuth)
	assert.Nil(err)

	res, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("my-apikey", *res.Key)

	res.Key = kong.String("my-apikey2")
	err = collection.Update(*res)
	assert.Nil(err)

	res, err = collection.Get("my-apikey")
	assert.NotNil(err)

	res, err = collection.Get("my-apikey2")
	assert.Nil(err)
	assert.Equal("first", *res.ID)
}

func TestKeyAuthDelete(t *testing.T) {
	assert := assert.New(t)
	collection := keyAuthsCollection()

	var keyAuth KeyAuth
	keyAuth.Key = kong.String("my-apikey1")
	keyAuth.ID = kong.String("first")
	keyAuth.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}
	err := collection.Add(keyAuth)
	assert.Nil(err)

	res, err := collection.Get("my-apikey1")
	assert.Nil(err)
	assert.NotNil(res)

	err = collection.Delete(*res.ID)
	assert.Nil(err)

	res, err = collection.Get("my-apikey1")
	assert.NotNil(err)
	assert.Nil(res)

	// delete a non-existing one
	err = collection.Delete("first")
	assert.NotNil(err)

	err = collection.Delete("my-apikey1")
	assert.NotNil(err)
}

func TestKeyAuthGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := keyAuthsCollection()

	populateWithKeyAuthFixtures(assert, collection)

	keyAuths, err := collection.GetAll()
	assert.Nil(err)
	assert.Equal(5, len(keyAuths))
}

func TestKeyAuthGetByConsumer(t *testing.T) {
	assert := assert.New(t)
	collection := keyAuthsCollection()

	populateWithKeyAuthFixtures(assert, collection)

	keyAuths, err := collection.GetAllByConsumerID("consumer1-id")
	assert.Nil(err)
	assert.Equal(3, len(keyAuths))

	keyAuths, err = collection.GetAllByConsumerUsername("consumer2-name")
	assert.Nil(err)
	assert.Equal(2, len(keyAuths))
}

func populateWithKeyAuthFixtures(assert *assert.Assertions,
	collection *KeyAuthsCollection) {

	keyAuths := []KeyAuth{
		{
			KeyAuth: kong.KeyAuth{
				Key: kong.String("my-apikey11"),
				ID:  kong.String("first"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			KeyAuth: kong.KeyAuth{
				Key: kong.String("my-apikey12"),
				ID:  kong.String("second"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			KeyAuth: kong.KeyAuth{
				Key: kong.String("my-apikey13"),
				ID:  kong.String("third"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			KeyAuth: kong.KeyAuth{
				Key: kong.String("my-apikey21"),
				ID:  kong.String("fourth"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer2-id"),
					Username: kong.String("consumer2-name"),
				},
			},
		},
		{
			KeyAuth: kong.KeyAuth{
				Key: kong.String("my-apikey22"),
				ID:  kong.String("fifth"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer2-id"),
					Username: kong.String("consumer2-name"),
				},
			},
		},
	}

	for _, k := range keyAuths {
		err := collection.Add(k)
		assert.Nil(err)
	}
}
