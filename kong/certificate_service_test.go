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

func TestCertificateListEndpoint(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// fixtures
	certificates := []*Certificate{
		{
			Cert: String("foo1"),
			Key:  String("bar1"),
		},
		{
			Cert: String("foo2"),
			Key:  String("bar2"),
		},
		{
			Cert: String("foo3"),
			Key:  String("bar3"),
		},
	}

	// create fixturs
	for i := 0; i < len(certificates); i++ {
		certificate, err := client.Certificates.Create(defaultCtx, certificates[i])
		assert.Nil(err)
		assert.NotNil(certificate)
		certificates[i] = certificate
	}

	certificatesFromKong, next, err := client.Certificates.List(defaultCtx, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(certificatesFromKong)
	assert.Equal(3, len(certificatesFromKong))

	// check if we see all certificates
	assert.True(compareCertificates(certificates, certificatesFromKong))

	// Test pagination
	certificatesFromKong = []*Certificate{}

	// first page
	page1, next, err := client.Certificates.List(defaultCtx, &ListOpt{Size: 1})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))
	certificatesFromKong = append(certificatesFromKong, page1...)

	// last page
	next.Size = 2
	page2, next, err := client.Certificates.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(page2)
	assert.Equal(2, len(page2))
	certificatesFromKong = append(certificatesFromKong, page2...)

	assert.True(compareCertificates(certificates, certificatesFromKong))

	for i := 0; i < len(certificates); i++ {
		assert.Nil(client.Certificates.Delete(defaultCtx, certificates[i].ID))
	}
}

func compareCertificates(expected, actual []*Certificate) bool {
	var expectedUsernames, actualUsernames []string
	for _, certificate := range expected {
		expectedUsernames = append(expectedUsernames, *certificate.Cert)
	}

	for _, certificate := range actual {
		actualUsernames = append(actualUsernames, *certificate.Cert)
	}

	return (compareSlices(expectedUsernames, actualUsernames))
}
