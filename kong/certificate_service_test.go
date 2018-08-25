package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestCertificatesService(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	certificate := &Certificate{
		Key:  String("foo"),
		Cert: String("bar"),
		SNIs: StringSlice("host1.com", "host2.com"),
	}

	createdCertificate, err := client.Certificates.Create(defaultCtx, certificate)
	assert.Nil(err)
	assert.NotNil(createdCertificate)

	certificate, err = client.Certificates.Get(defaultCtx, createdCertificate.ID)
	assert.Nil(err)
	assert.NotNil(certificate)
	assert.Equal(2, len(createdCertificate.SNIs))

	certificate.Key = String("baz")
	certificate, err = client.Certificates.Update(defaultCtx, certificate)
	assert.Nil(err)
	assert.NotNil(certificate)
	assert.Equal("baz", *certificate.Key)

	err = client.Certificates.Delete(defaultCtx, createdCertificate.ID)
	assert.Nil(err)

	// ID can be specified
	id := uuid.NewV4().String()
	certificate = &Certificate{
		Key:  String("key"),
		Cert: String("cert"),
		ID:   String(id),
	}

	createdCertificate, err = client.Certificates.Create(defaultCtx, certificate)
	assert.Nil(err)
	assert.NotNil(createdCertificate)
	assert.Equal(id, *createdCertificate.ID)

	err = client.Certificates.Delete(defaultCtx, createdCertificate.ID)
	assert.Nil(err)
}
