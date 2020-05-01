package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func jwtAuthsCollection() *JWTAuthsCollection {
	return state().JWTAuths
}

func TestJWTAuthInsert(t *testing.T) {
	assert := assert.New(t)
	collection := jwtAuthsCollection()

	var jwtAuth JWTAuth
	jwtAuth.Key = kong.String("my-key")
	jwtAuth.ID = kong.String("first")
	err := collection.Add(jwtAuth)
	assert.NotNil(err)

	var jwtAuth2 JWTAuth
	jwtAuth2.Key = kong.String("my-key")
	jwtAuth2.ID = kong.String("first")
	jwtAuth2.Consumer = &kong.Consumer{
		ID:       kong.String("consumer-id"),
		Username: kong.String("my-username"),
	}
	err = collection.Add(jwtAuth2)
	assert.Nil(err)
}

func TestJWTAuthGet(t *testing.T) {
	assert := assert.New(t)
	collection := jwtAuthsCollection()

	var jwtAuth JWTAuth
	jwtAuth.Key = kong.String("my-key")
	jwtAuth.ID = kong.String("first")
	jwtAuth.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}

	err := collection.Add(jwtAuth)
	assert.Nil(err)

	res, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("my-key", *res.Key)

	res, err = collection.Get("my-key")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("first", *res.ID)
	assert.Equal("consumer1-id", *res.Consumer.ID)

	res, err = collection.Get("does-not-exist")
	assert.NotNil(err)
	assert.Nil(res)
}

func TestJWTAuthUpdate(t *testing.T) {
	assert := assert.New(t)
	collection := jwtAuthsCollection()

	var jwtAuth JWTAuth
	jwtAuth.Key = kong.String("my-key")
	jwtAuth.ID = kong.String("first")
	jwtAuth.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}

	err := collection.Add(jwtAuth)
	assert.Nil(err)

	res, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("my-key", *res.Key)

	res.Key = kong.String("my-key2")
	err = collection.Update(*res)
	assert.Nil(err)

	res, err = collection.Get("my-key")
	assert.NotNil(err)
	assert.Nil(res)

	res, err = collection.Get("my-key2")
	assert.Nil(err)
	assert.Equal("first", *res.ID)
}

func TestJWTAuthDelete(t *testing.T) {
	assert := assert.New(t)
	collection := jwtAuthsCollection()

	var jwtAuth JWTAuth
	jwtAuth.Key = kong.String("my-key1")
	jwtAuth.ID = kong.String("first")
	jwtAuth.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}
	err := collection.Add(jwtAuth)
	assert.Nil(err)

	res, err := collection.Get("my-key1")
	assert.Nil(err)
	assert.NotNil(res)

	err = collection.Delete(*res.ID)
	assert.Nil(err)

	res, err = collection.Get("my-key1")
	assert.NotNil(err)
	assert.Nil(res)

	// delete a non-existing one
	err = collection.Delete("first")
	assert.NotNil(err)

	err = collection.Delete("my-key1")
	assert.NotNil(err)
}

func TestJWTAuthGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := jwtAuthsCollection()

	populateWithJWTAuthFixtures(assert, collection)

	jwtAuths, err := collection.GetAll()
	assert.Nil(err)
	assert.Equal(5, len(jwtAuths))
}

func TestJWTAuthGetByConsumer(t *testing.T) {
	assert := assert.New(t)
	collection := jwtAuthsCollection()

	populateWithJWTAuthFixtures(assert, collection)

	jwtAuths, err := collection.GetAllByConsumerID("consumer1-id")
	assert.Nil(err)
	assert.Equal(3, len(jwtAuths))
}

func populateWithJWTAuthFixtures(assert *assert.Assertions,
	collection *JWTAuthsCollection) {

	jwtAuths := []JWTAuth{
		{
			JWTAuth: kong.JWTAuth{
				Key: kong.String("my-key11"),
				ID:  kong.String("first"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			JWTAuth: kong.JWTAuth{
				Key: kong.String("my-key12"),
				ID:  kong.String("second"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			JWTAuth: kong.JWTAuth{
				Key: kong.String("my-key13"),
				ID:  kong.String("third"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			JWTAuth: kong.JWTAuth{
				Key: kong.String("my-key21"),
				ID:  kong.String("fourth"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer2-id"),
					Username: kong.String("consumer2-name"),
				},
			},
		},
		{
			JWTAuth: kong.JWTAuth{
				Key: kong.String("my-key22"),
				ID:  kong.String("fifth"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer2-id"),
					Username: kong.String("consumer2-name"),
				},
			},
		},
	}

	for _, k := range jwtAuths {
		err := collection.Add(k)
		assert.Nil(err)
	}
}
