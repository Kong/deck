package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestSNIsService(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	sni := &SNI{
		Name: String("host1.com"),
	}

	// cert is required
	badSNI, err := client.SNIs.Create(defaultCtx, sni)
	assert.NotNil(err)
	assert.Nil(badSNI)

	// create a cert
	fixtureCertificate, err := client.Certificates.Create(defaultCtx, &Certificate{
		Key:  String("foo"),
		Cert: String("bar"),
	})
	assert.Nil(err)
	assert.NotNil(fixtureCertificate)
	assert.NotNil(fixtureCertificate.ID)

	createdSNI, err := client.SNIs.Create(defaultCtx, &SNI{
		Name:        String("host1.com"),
		Certificate: fixtureCertificate,
	})
	assert.Nil(err)
	assert.NotNil(createdSNI)

	sni, err = client.SNIs.Get(defaultCtx, createdSNI.ID)
	assert.Nil(err)
	assert.NotNil(sni)

	sni.Name = String("host2.com")
	sni, err = client.SNIs.Update(defaultCtx, sni)
	assert.Nil(err)
	assert.NotNil(sni)
	assert.Equal("host2.com", *sni.Name)

	err = client.SNIs.Delete(defaultCtx, createdSNI.ID)
	assert.Nil(err)

	// ID can be specified
	id := uuid.NewV4().String()
	sni = &SNI{
		Name:        String("host3.com"),
		ID:          String(id),
		Certificate: fixtureCertificate,
	}

	createdSNI, err = client.SNIs.Create(defaultCtx, sni)
	assert.Nil(err)
	assert.NotNil(createdSNI)
	assert.Equal(id, *createdSNI.ID)

	err = client.Certificates.Delete(defaultCtx, fixtureCertificate.ID)
	assert.Nil(err)
}
