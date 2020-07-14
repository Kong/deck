// +build enterprise

package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestMTLSCreate(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	mtls, err := client.MTLSAuths.Create(defaultCtx, String("foo"), nil)
	assert.NotNil(err)
	assert.Nil(mtls)

	mtls = &MTLSAuth{}
	mtls, err = client.MTLSAuths.Create(defaultCtx, String(""), mtls)
	assert.NotNil(err)
	assert.Nil(mtls)

	mtls, err = client.MTLSAuths.Create(defaultCtx,
		String("does-not-exist"), mtls)
	assert.NotNil(err)
	assert.Nil(mtls)

	// consumer for the MTLS
	consumer := &Consumer{
		Username: String("foo"),
	}

	// without a CA certificate attached
	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	mtls = &MTLSAuth{
		SubjectName: String("test@example.com"),
	}
	mtls, err = client.MTLSAuths.Create(defaultCtx, consumer.ID, mtls)
	assert.Nil(err)
	assert.NotNil(mtls)
	assert.Equal("test@example.com", *mtls.SubjectName)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))

	// with a CA certificate attached
	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	certificate := &CACertificate{
		Cert: String(caCert1),
	}
	createdCertificate, err := client.CACertificates.Create(defaultCtx,
		certificate)
	assert.Nil(err)

	assert.NotNil(createdCertificate)
	mtls = &MTLSAuth{
		SubjectName:   String("test@example.com"),
		CACertificate: createdCertificate,
	}
	mtls, err = client.MTLSAuths.Create(defaultCtx, consumer.ID, mtls)
	assert.Nil(err)
	assert.NotNil(mtls)
	assert.Equal("test@example.com", *mtls.SubjectName)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestMTLSCreateWithID(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	mtls := &MTLSAuth{
		ID:          String(uuid),
		SubjectName: String("test@example.com"),
	}

	// consumer for the mtls
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdMTLS, err := client.MTLSAuths.Create(defaultCtx, consumer.ID,
		mtls)
	assert.Nil(err)
	assert.NotNil(createdMTLS)

	assert.Equal(uuid, *createdMTLS.ID)
	assert.Equal("test@example.com", *mtls.SubjectName)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestMTLSGet(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	mtls := &MTLSAuth{
		ID:          String(uuid),
		SubjectName: String("test@example.com"),
	}

	// consumer for the mtls
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdMTLS, err := client.MTLSAuths.Create(defaultCtx, consumer.ID, mtls)
	assert.Nil(err)
	assert.NotNil(createdMTLS)

	mtls, err = client.MTLSAuths.Get(defaultCtx, consumer.ID, mtls.ID)
	assert.Nil(err)
	assert.Equal("test@example.com", *mtls.SubjectName)

	mtls, err = client.MTLSAuths.Get(defaultCtx, consumer.ID,
		String("does-not-exists"))
	assert.Nil(mtls)
	assert.NotNil(err)

	mtls, err = client.MTLSAuths.Get(defaultCtx, consumer.ID, String(""))
	assert.Nil(mtls)
	assert.NotNil(err)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestMTLSUpdate(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	mtls := &MTLSAuth{
		ID:          String(uuid),
		SubjectName: String("test@example.com"),
	}

	// consumer for the mtls
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdMTLS, err := client.MTLSAuths.Create(defaultCtx, consumer.ID, mtls)
	assert.Nil(err)
	assert.NotNil(createdMTLS)

	mtls, err = client.MTLSAuths.Get(defaultCtx, consumer.ID, mtls.ID)
	assert.Nil(err)
	assert.Equal("test@example.com", *mtls.SubjectName)

	mtls.SubjectName = String("different@example.com")
	updatedMTLS, err := client.MTLSAuths.Update(defaultCtx, consumer.ID, mtls)
	assert.Nil(err)
	assert.NotNil(updatedMTLS)
	assert.Equal("different@example.com", *updatedMTLS.SubjectName)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestMTLSDelete(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	mtls := &MTLSAuth{
		ID:          String(uuid),
		SubjectName: String("test@example.com"),
	}

	// consumer for the mtls
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdMTLS, err := client.MTLSAuths.Create(defaultCtx, consumer.ID, mtls)
	assert.Nil(err)
	assert.NotNil(createdMTLS)

	err = client.MTLSAuths.Delete(defaultCtx, consumer.ID, mtls.ID)
	assert.Nil(err)

	mtls, err = client.MTLSAuths.Get(defaultCtx, consumer.ID, mtls.ID)
	assert.NotNil(err)
	assert.Nil(mtls)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestMTLSListMethods(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// consumer for the MTLS
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
	mtlss := []*MTLSAuth{
		{
			SubjectName: String("username11@example.com"),
			Consumer:    consumer1,
		},
		{
			SubjectName: String("username12@example.com"),
			Consumer:    consumer1,
		},
		{
			SubjectName: String("username21@example.com"),
			Consumer:    consumer2,
		},
		{
			SubjectName: String("username22@example.com"),
			Consumer:    consumer2,
		},
	}

	// create fixturs
	for i := 0; i < len(mtlss); i++ {
		mtls, err := client.MTLSAuths.Create(defaultCtx,
			mtlss[i].Consumer.ID, mtlss[i])
		assert.Nil(err)
		assert.NotNil(mtls)
		mtlss[i] = mtls
	}

	mtlssFromKong, next, err := client.MTLSAuths.List(defaultCtx, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(mtlssFromKong)
	assert.Equal(4, len(mtlssFromKong))

	// first page
	page1, next, err := client.MTLSAuths.List(defaultCtx, &ListOpt{Size: 1})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))

	// last page
	next.Size = 3
	page2, next, err := client.MTLSAuths.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(page2)
	assert.Equal(3, len(page2))

	mtlssForConsumer, next, err := client.MTLSAuths.ListForConsumer(defaultCtx,
		consumer1.ID, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(mtlssForConsumer)
	assert.Equal(2, len(mtlssForConsumer))

	mtlss, err = client.MTLSAuths.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(mtlss)
	assert.Equal(4, len(mtlss))

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer1.ID))
	assert.Nil(client.Consumers.Delete(defaultCtx, consumer2.ID))
}
