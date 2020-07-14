package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWTCreate(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	jwt, err := client.JWTAuths.Create(defaultCtx, String("foo"), nil)
	assert.NotNil(err)
	assert.Nil(jwt)

	jwt = &JWTAuth{}
	jwt, err = client.JWTAuths.Create(defaultCtx, String(""), jwt)
	assert.NotNil(err)
	assert.Nil(jwt)

	jwt, err = client.JWTAuths.Create(defaultCtx,
		String("does-not-exist"), jwt)
	assert.NotNil(err)
	assert.Nil(jwt)

	// consumer for the JWT
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	jwt = &JWTAuth{
		Key:          String("foo"),
		RSAPublicKey: String("bar"),
	}
	jwt, err = client.JWTAuths.Create(defaultCtx, consumer.ID, jwt)
	assert.Nil(err)
	assert.NotNil(jwt)
	assert.NotEmpty(*jwt.Secret)
	assert.Equal("bar", *jwt.RSAPublicKey)
	assert.NotEmpty(*jwt.Algorithm)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestJWTCreateWithID(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	jwt := &JWTAuth{
		ID:     String(uuid),
		Key:    String("my-key"),
		Secret: String("my-secret"),
	}

	// consumer for the jwt
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdJWT, err := client.JWTAuths.Create(defaultCtx, consumer.ID,
		jwt)
	assert.Nil(err)
	assert.NotNil(createdJWT)

	assert.Equal(uuid, *createdJWT.ID)
	assert.Equal("my-key", *createdJWT.Key)
	assert.Equal("my-secret", *createdJWT.Secret)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestJWTGet(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	jwt := &JWTAuth{
		ID:  String(uuid),
		Key: String("my-key"),
	}

	// consumer for the jwt
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdJWT, err := client.JWTAuths.Create(defaultCtx, consumer.ID, jwt)
	assert.Nil(err)
	assert.NotNil(createdJWT)

	jwt, err = client.JWTAuths.Get(defaultCtx, consumer.ID, jwt.ID)
	assert.Nil(err)
	assert.Equal("my-key", *jwt.Key)

	jwt, err = client.JWTAuths.Get(defaultCtx, consumer.ID,
		jwt.Key)
	assert.Nil(err)
	assert.Equal("my-key", *jwt.Key)

	jwt, err = client.JWTAuths.Get(defaultCtx, consumer.ID,
		String("does-not-exists"))
	assert.Nil(jwt)
	assert.NotNil(err)

	jwt, err = client.JWTAuths.Get(defaultCtx, consumer.ID, String(""))
	assert.Nil(jwt)
	assert.NotNil(err)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestJWTUpdate(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	jwt := &JWTAuth{
		ID:  String(uuid),
		Key: String("my-key"),
	}

	// consumer for the jwt
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdJWT, err := client.JWTAuths.Create(defaultCtx, consumer.ID, jwt)
	assert.Nil(err)
	assert.NotNil(createdJWT)

	jwt, err = client.JWTAuths.Get(defaultCtx, consumer.ID, jwt.ID)
	assert.Nil(err)
	assert.Equal("my-key", *jwt.Key)

	jwt.Key = String("my-new-key")
	jwt.Secret = String("my-new-secret")
	updatedJWT, err := client.JWTAuths.Update(defaultCtx, consumer.ID, jwt)
	assert.Nil(err)
	assert.NotNil(updatedJWT)
	assert.Equal("my-new-secret", *updatedJWT.Secret)
	assert.Equal("my-new-key", *updatedJWT.Key)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestJWTDelete(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	jwt := &JWTAuth{
		ID:  String(uuid),
		Key: String("my-key"),
	}

	// consumer for the jwt
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdJWT, err := client.JWTAuths.Create(defaultCtx, consumer.ID, jwt)
	assert.Nil(err)
	assert.NotNil(createdJWT)

	err = client.JWTAuths.Delete(defaultCtx, consumer.ID, jwt.Key)
	assert.Nil(err)

	jwt, err = client.JWTAuths.Get(defaultCtx, consumer.ID, jwt.Key)
	assert.NotNil(err)
	assert.Nil(jwt)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestJWTListMethods(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// consumer for the JWT
	consumer1 := &Consumer{
		Username: String("foo"),
	}

	consumer1, err = client.Consumers.Create(defaultCtx, consumer1)
	assert.Nil(err)
	assert.NotNil(consumer1)

	consumer2 := &Consumer{
		Username: String("bar"),
	}

	consumer2, err = client.Consumers.Create(defaultCtx, consumer2)
	assert.Nil(err)
	assert.NotNil(consumer2)

	// fixtures
	jwts := []*JWTAuth{
		{
			Key:      String("username11"),
			Consumer: consumer1,
		},
		{
			Key:      String("username12"),
			Consumer: consumer1,
		},
		{
			Key:      String("username21"),
			Consumer: consumer2,
		},
		{
			Key:      String("username22"),
			Consumer: consumer2,
		},
	}

	// create fixturs
	for i := 0; i < len(jwts); i++ {
		jwt, err := client.JWTAuths.Create(defaultCtx,
			jwts[i].Consumer.ID, jwts[i])
		assert.Nil(err)
		assert.NotNil(jwt)
		jwts[i] = jwt
	}

	jwtsFromKong, next, err := client.JWTAuths.List(defaultCtx, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(jwtsFromKong)
	assert.Equal(4, len(jwtsFromKong))

	// first page
	page1, next, err := client.JWTAuths.List(defaultCtx, &ListOpt{Size: 1})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))

	// last page
	next.Size = 3
	page2, next, err := client.JWTAuths.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(page2)
	assert.Equal(3, len(page2))

	jwtsForConsumer, next, err := client.JWTAuths.ListForConsumer(defaultCtx,
		consumer1.ID, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(jwtsForConsumer)
	assert.Equal(2, len(jwtsForConsumer))

	jwts, err = client.JWTAuths.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(jwts)
	assert.Equal(4, len(jwts))

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer1.ID))
	assert.Nil(client.Consumers.Delete(defaultCtx, consumer2.ID))
}
