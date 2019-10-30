package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func hmacAuthsCollection() *HMACAuthsCollection {
	return state().HMACAuths
}

func TestHMACAuthInsert(t *testing.T) {
	assert := assert.New(t)
	collection := hmacAuthsCollection()

	var hmacAuth HMACAuth
	hmacAuth.ID = kong.String("first")
	err := collection.Add(hmacAuth)
	assert.NotNil(err)

	hmacAuth.Username = kong.String("my-username")
	err = collection.Add(hmacAuth)
	assert.NotNil(err)

	var hmacAuth2 HMACAuth
	hmacAuth2.Username = kong.String("my-username")
	hmacAuth2.ID = kong.String("first")
	hmacAuth2.Consumer = &kong.Consumer{
		ID:       kong.String("consumer-id"),
		Username: kong.String("my-username"),
	}
	err = collection.Add(hmacAuth2)
	assert.Nil(err)
}

func TestHMACAuthGet(t *testing.T) {
	assert := assert.New(t)
	collection := hmacAuthsCollection()

	var hmacAuth HMACAuth
	hmacAuth.Username = kong.String("my-username")
	hmacAuth.ID = kong.String("first")
	hmacAuth.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}

	err := collection.Add(hmacAuth)
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

func TestHMACAuthUpdate(t *testing.T) {
	assert := assert.New(t)
	collection := hmacAuthsCollection()

	var hmacAuth HMACAuth
	hmacAuth.Username = kong.String("my-username")
	hmacAuth.ID = kong.String("first")
	hmacAuth.Consumer = &kong.Consumer{
		ID: kong.String("consumer1-id"),
	}

	err := collection.Add(hmacAuth)
	assert.Nil(err)

	res, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("my-username", *res.Username)

	res.Username = kong.String("my-username2")
	res.Secret = kong.String("secret")
	err = collection.Update(*res)
	assert.Nil(err)

	res, err = collection.Get("my-username")
	assert.NotNil(err)

	res, err = collection.Get("my-username2")
	assert.Nil(err)
	assert.Equal("first", *res.ID)
	assert.Equal("secret", *res.Secret)
}

func TestHMACAuthDelete(t *testing.T) {
	assert := assert.New(t)
	collection := hmacAuthsCollection()

	var hmacAuth HMACAuth
	hmacAuth.Username = kong.String("my-username1")
	hmacAuth.ID = kong.String("first")
	hmacAuth.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}
	err := collection.Add(hmacAuth)
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

func TestHMACAuthGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := hmacAuthsCollection()

	populateWithHMACAuthFixtures(assert, collection)

	hmacAuths, err := collection.GetAll()
	assert.Nil(err)
	assert.Equal(5, len(hmacAuths))
}

func TestHMACAuthGetByConsumer(t *testing.T) {
	assert := assert.New(t)
	collection := hmacAuthsCollection()

	populateWithHMACAuthFixtures(assert, collection)

	hmacAuths, err := collection.GetAllByConsumerID("consumer1-id")
	assert.Nil(err)
	assert.Equal(3, len(hmacAuths))
}

func populateWithHMACAuthFixtures(assert *assert.Assertions,
	collection *HMACAuthsCollection) {

	hmacAuths := []HMACAuth{
		{
			HMACAuth: kong.HMACAuth{
				Username: kong.String("my-username11"),
				ID:       kong.String("first"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			HMACAuth: kong.HMACAuth{
				Username: kong.String("my-username12"),
				ID:       kong.String("second"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			HMACAuth: kong.HMACAuth{
				Username: kong.String("my-username13"),
				ID:       kong.String("third"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			HMACAuth: kong.HMACAuth{
				Username: kong.String("my-username21"),
				ID:       kong.String("fourth"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer2-id"),
					Username: kong.String("consumer2-name"),
				},
			},
		},
		{
			HMACAuth: kong.HMACAuth{
				Username: kong.String("my-username22"),
				ID:       kong.String("fifth"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer2-id"),
					Username: kong.String("consumer2-name"),
				},
			},
		},
	}

	for _, k := range hmacAuths {
		err := collection.Add(k)
		assert.Nil(err)
	}
}
