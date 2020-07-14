package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestOauth2CredentialCreate(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	oauth2Cred, err := client.Oauth2Credentials.Create(defaultCtx,
		String("foo"), nil)
	assert.NotNil(err)
	assert.Nil(oauth2Cred)

	oauth2Cred = &Oauth2Credential{}
	oauth2Cred, err = client.Oauth2Credentials.Create(defaultCtx, String(""),
		oauth2Cred)
	assert.NotNil(err)
	assert.Nil(oauth2Cred)

	// consumer for the oauth2 cred
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	oauth2Cred = &Oauth2Credential{
		ClientID:     String("foo"),
		Name:         String("name-foo"),
		RedirectURIs: StringSlice("http://foo.com", "http://bar.com"),
	}
	oauth2Cred, err = client.Oauth2Credentials.Create(defaultCtx,
		consumer.ID, oauth2Cred)
	assert.Nil(err)
	assert.NotNil(oauth2Cred)
	assert.NotNil(oauth2Cred.ClientSecret)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestOauth2CredentialCreateWithID(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	oauth2Cred := &Oauth2Credential{
		ID:           String(uuid),
		Name:         String("name"),
		ClientSecret: String("my-client-secret"),
		RedirectURIs: StringSlice("http://foo.com", "http://bar.com"),
		ClientID:     String("my-clientid"),
	}

	// consumer for the oauth2 cred
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdOauth2Credential, err := client.Oauth2Credentials.Create(
		defaultCtx, consumer.ID, oauth2Cred)
	assert.Nil(err)
	assert.NotNil(createdOauth2Credential)

	assert.Equal(uuid, *createdOauth2Credential.ID)
	assert.Equal("my-clientid", *createdOauth2Credential.ClientID)
	assert.Equal("my-client-secret", *createdOauth2Credential.ClientSecret)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestOauth2CredentialGet(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	oauth2Cred := &Oauth2Credential{
		ID:           String(uuid),
		Name:         String("name-foo"),
		ClientID:     String("foo-clientid"),
		RedirectURIs: StringSlice("http://foo.com", "http://bar.com"),
	}

	// consumer for the oauth2 cred
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdOauth2Credential, err := client.Oauth2Credentials.Create(defaultCtx,
		consumer.ID, oauth2Cred)
	assert.Nil(err)
	assert.NotNil(createdOauth2Credential)

	oauth2Cred, err = client.Oauth2Credentials.Get(defaultCtx,
		consumer.ID, oauth2Cred.ID)
	assert.Nil(err)
	assert.Equal("foo-clientid", *oauth2Cred.ClientID)

	oauth2Cred, err = client.Oauth2Credentials.Get(defaultCtx, consumer.ID,
		String("foo-clientid"))
	assert.Nil(err)
	assert.NotNil(oauth2Cred)

	oauth2Cred, err = client.Oauth2Credentials.Get(defaultCtx, consumer.ID,
		String("does-not-exists"))
	assert.Nil(oauth2Cred)
	assert.NotNil(err)

	oauth2Cred, err = client.Oauth2Credentials.Get(defaultCtx,
		consumer.ID, String(""))
	assert.Nil(oauth2Cred)
	assert.NotNil(err)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestOauth2CredentialUpdate(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	oauth2Cred := &Oauth2Credential{
		ID:           String(uuid),
		ClientID:     String("client-id"),
		Name:         String("foo-name"),
		RedirectURIs: StringSlice("http://foo.com", "http://bar.com"),
	}

	// consumer for the oauth2 cred
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdOauth2Credential, err := client.Oauth2Credentials.Create(
		defaultCtx, consumer.ID, oauth2Cred)
	assert.Nil(err)
	assert.NotNil(createdOauth2Credential)

	oauth2Cred, err = client.Oauth2Credentials.Get(defaultCtx,
		consumer.ID, oauth2Cred.ID)
	assert.Nil(err)
	assert.Equal("foo-name", *oauth2Cred.Name)

	oauth2Cred.Name = String("new-foo-name")
	oauth2Cred.ClientSecret = String("my-new-secret")
	updatedOauth2Credential, err :=
		client.Oauth2Credentials.Update(defaultCtx, consumer.ID, oauth2Cred)
	assert.Nil(err)
	assert.NotNil(updatedOauth2Credential)
	assert.Equal("new-foo-name", *updatedOauth2Credential.Name)
	assert.Equal("my-new-secret", *updatedOauth2Credential.ClientSecret)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestOauth2CredentialDelete(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	oauth2Cred := &Oauth2Credential{
		ID:           String(uuid),
		ClientID:     String("my-client-id"),
		Name:         String("my-name"),
		RedirectURIs: StringSlice("http://foo.com", "http://bar.com"),
	}

	// consumer for the oauth2 cred
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdOauth2Credential, err :=
		client.Oauth2Credentials.Create(defaultCtx, consumer.ID, oauth2Cred)
	assert.Nil(err)
	assert.NotNil(createdOauth2Credential)

	err = client.Oauth2Credentials.Delete(defaultCtx,
		consumer.ID, oauth2Cred.ClientID)
	assert.Nil(err)

	oauth2Cred, err = client.Oauth2Credentials.Get(defaultCtx,
		consumer.ID, oauth2Cred.ClientID)
	assert.NotNil(err)
	assert.Nil(oauth2Cred)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestOauth2CredentialListMethods(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// consumer for the oauth2 cred
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
	oauth2Creds := []*Oauth2Credential{
		{
			ClientID:     String("clientid11"),
			Name:         String("name11"),
			RedirectURIs: StringSlice("http://foo.com", "http://bar.com"),
			Consumer:     consumer1,
		},
		{
			ClientID:     String("clientid12"),
			Name:         String("name12"),
			RedirectURIs: StringSlice("http://foo.com", "http://bar.com"),
			Consumer:     consumer1,
		},
		{
			ClientID:     String("clientid21"),
			Name:         String("name21"),
			RedirectURIs: StringSlice("http://foo.com", "http://bar.com"),
			Consumer:     consumer2,
		},
		{
			ClientID:     String("clientid22"),
			Name:         String("name22"),
			RedirectURIs: StringSlice("http://foo.com", "http://bar.com"),
			Consumer:     consumer2,
		},
	}

	// create fixturs
	for i := 0; i < len(oauth2Creds); i++ {
		oauth2Cred, err := client.Oauth2Credentials.Create(defaultCtx,
			oauth2Creds[i].Consumer.ID, oauth2Creds[i])
		assert.Nil(err)
		assert.NotNil(oauth2Cred)
		oauth2Creds[i] = oauth2Cred
	}

	oauth2CredsFromKong, next, err :=
		client.Oauth2Credentials.List(defaultCtx, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(oauth2CredsFromKong)
	assert.Equal(4, len(oauth2CredsFromKong))

	// first page
	page1, next, err :=
		client.Oauth2Credentials.List(defaultCtx, &ListOpt{Size: 1})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))

	// last page
	next.Size = 3
	page2, next, err := client.Oauth2Credentials.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(page2)
	assert.Equal(3, len(page2))

	oauth2CredsForConsumer, next, err :=
		client.Oauth2Credentials.ListForConsumer(defaultCtx, consumer1.ID, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(oauth2CredsForConsumer)
	assert.Equal(2, len(oauth2CredsForConsumer))

	oauth2Creds, err = client.Oauth2Credentials.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(oauth2Creds)
	assert.Equal(4, len(oauth2Creds))

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer1.ID))
	assert.Nil(client.Consumers.Delete(defaultCtx, consumer2.ID))
}
