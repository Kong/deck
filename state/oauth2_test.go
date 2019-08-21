package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func TestOauth2CredInsert(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewOauth2CredsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var oauth2Cred Oauth2Credential
	oauth2Cred.ClientID = kong.String("client-id")
	oauth2Cred.ID = kong.String("first")
	err = collection.Add(oauth2Cred)
	assert.NotNil(err)

	oauth2Cred.Consumer = &kong.Consumer{
		ID:       kong.String("consumer-id"),
		Username: kong.String("my-username"),
	}
	err = collection.Add(oauth2Cred)
	assert.Nil(err)
}

func TestOauth2CredentialGet(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewOauth2CredsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var oauth2Cred Oauth2Credential
	oauth2Cred.ClientID = kong.String("my-clientid")
	oauth2Cred.ID = kong.String("first")
	oauth2Cred.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}

	err = collection.Add(oauth2Cred)
	assert.Nil(err)

	res, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("my-clientid", *res.ClientID)

	res, err = collection.Get("my-clientid")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("first", *res.ID)
	assert.Equal("consumer1-id", *res.Consumer.ID)

	res, err = collection.Get("does-not-exist")
	assert.NotNil(err)
	assert.Nil(res)
}

func TestOauth2CredentialUpdate(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewOauth2CredsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var oauth2Cred Oauth2Credential
	oauth2Cred.ClientID = kong.String("my-clientid")
	oauth2Cred.ID = kong.String("first")
	oauth2Cred.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}

	err = collection.Add(oauth2Cred)
	assert.Nil(err)

	res, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("my-clientid", *res.ClientID)

	res.ClientID = kong.String("my-clientid2")
	err = collection.Update(*res)
	assert.Nil(err)

	res, err = collection.Get("my-clientid")
	assert.NotNil(err)

	res, err = collection.Get("my-clientid2")
	assert.Nil(err)
	assert.Equal("first", *res.ID)
}

func TestOauth2CredentialDelete(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewOauth2CredsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var oauth2Cred Oauth2Credential
	oauth2Cred.ClientID = kong.String("my-clientid1")
	oauth2Cred.ID = kong.String("first")
	oauth2Cred.Consumer = &kong.Consumer{
		ID:       kong.String("consumer1-id"),
		Username: kong.String("consumer1-name"),
	}
	err = collection.Add(oauth2Cred)
	assert.Nil(err)

	res, err := collection.Get("my-clientid1")
	assert.Nil(err)
	assert.NotNil(res)

	err = collection.Delete(*res.ID)
	assert.Nil(err)

	res, err = collection.Get("my-clientid1")
	assert.NotNil(err)
	assert.Nil(res)

	// delete a non-existing one
	err = collection.Delete("first")
	assert.NotNil(err)

	err = collection.Delete("my-clientid1")
	assert.NotNil(err)
}

func TestOauth2CredentialGetAll(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewOauth2CredsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	populateWithOauth2CredentialFixtures(assert, collection)

	oauth2Creds, err := collection.GetAll()
	assert.Nil(err)
	assert.Equal(5, len(oauth2Creds))
}

func TestOauth2CredentialGetByConsumer(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewOauth2CredsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	populateWithOauth2CredentialFixtures(assert, collection)

	oauth2Creds, err := collection.GetAllByConsumerID("consumer1-id")
	assert.Nil(err)
	assert.Equal(3, len(oauth2Creds))

	oauth2Creds, err = collection.GetAllByConsumerUsername("consumer2-name")
	assert.Nil(err)
	assert.Equal(2, len(oauth2Creds))
}

func populateWithOauth2CredentialFixtures(assert *assert.Assertions,
	collection *Oauth2CredsCollection) {

	oauth2Creds := []Oauth2Credential{
		{
			Oauth2Credential: kong.Oauth2Credential{
				ClientID: kong.String("my-clientid11"),
				ID:       kong.String("first"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			Oauth2Credential: kong.Oauth2Credential{
				ClientID: kong.String("my-clientid12"),
				ID:       kong.String("second"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			Oauth2Credential: kong.Oauth2Credential{
				ClientID: kong.String("my-clientid13"),
				ID:       kong.String("third"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer1-id"),
					Username: kong.String("consumer1-name"),
				},
			},
		},
		{
			Oauth2Credential: kong.Oauth2Credential{
				ClientID: kong.String("my-clientid21"),
				ID:       kong.String("fourth"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer2-id"),
					Username: kong.String("consumer2-name"),
				},
			},
		},
		{
			Oauth2Credential: kong.Oauth2Credential{
				ClientID: kong.String("my-clientid22"),
				ID:       kong.String("fifth"),
				Consumer: &kong.Consumer{
					ID:       kong.String("consumer2-id"),
					Username: kong.String("consumer2-name"),
				},
			},
		},
	}

	for _, k := range oauth2Creds {
		err := collection.Add(k)
		assert.Nil(err)
	}
}
