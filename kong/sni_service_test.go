package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestSNIsCertificate(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
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
	fixtureCertificate, err := client.Certificates.Create(defaultCtx,
		&Certificate{
			Key:  String(key1),
			Cert: String(cert1),
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

func TestSNIWithTags(T *testing.T) {
	runWhenKong(T, ">=1.1.0")
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	fixtureCertificate, err := client.Certificates.Create(defaultCtx,
		&Certificate{
			Key:  String(key1),
			Cert: String(cert1),
		})
	assert.Nil(err)

	createdSNI, err := client.SNIs.Create(defaultCtx, &SNI{
		Name:        String("host1.com"),
		Certificate: fixtureCertificate,
		Tags:        StringSlice("tag1", "tag2"),
	})
	assert.Nil(err)
	assert.NotNil(createdSNI)
	assert.Equal(StringSlice("tag1", "tag2"), createdSNI.Tags)

	err = client.Certificates.Delete(defaultCtx, fixtureCertificate.ID)
	assert.Nil(err)
}

func TestSNIListEndpoint(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	certificate := &Certificate{
		Cert: String(cert2),
		Key:  String(key2),
	}

	createdCertificate, err := client.Certificates.Create(defaultCtx,
		certificate)
	assert.Nil(err)
	assert.NotNil(createdCertificate)

	// fixtures
	snis := []*SNI{
		{
			Name:        String("sni1"),
			Certificate: createdCertificate,
		},
		{
			Name:        String("sni2"),
			Certificate: createdCertificate,
		},
		{
			Name:        String("sni3"),
			Certificate: createdCertificate,
		},
	}

	// create fixturs
	for i := 0; i < len(snis); i++ {
		sni, err := client.SNIs.Create(defaultCtx, snis[i])
		assert.Nil(err)
		assert.NotNil(sni)
		snis[i] = sni
	}

	snisFromKong, next, err := client.SNIs.List(defaultCtx, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(snisFromKong)
	assert.Equal(3, len(snisFromKong))

	// check if we see all snis
	assert.True(compareSNIs(snis, snisFromKong))

	// Test pagination
	snisFromKong = []*SNI{}

	// first page
	page1, next, err := client.SNIs.List(defaultCtx, &ListOpt{Size: 1})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))
	snisFromKong = append(snisFromKong, page1...)

	// last page
	next.Size = 2
	page2, next, err := client.SNIs.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(page2)
	assert.Equal(2, len(page2))
	snisFromKong = append(snisFromKong, page2...)

	assert.True(compareSNIs(snis, snisFromKong))

	snisForCert, next, err := client.SNIs.ListForCertificate(defaultCtx,
		createdCertificate.ID, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(snisForCert)

	assert.True(compareSNIs(snis, snisForCert))

	snis, err = client.SNIs.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(snis)
	assert.Equal(3, len(snis))

	for i := 0; i < len(snis); i++ {
		assert.Nil(client.SNIs.Delete(defaultCtx, snis[i].ID))
	}

	assert.Nil(client.Certificates.Delete(defaultCtx, createdCertificate.ID))
}

func compareSNIs(expected, actual []*SNI) bool {
	var expectedUsernames, actualUsernames []string
	for _, sni := range expected {
		expectedUsernames = append(expectedUsernames, *sni.Name)
	}

	for _, sni := range actual {
		actualUsernames = append(actualUsernames, *sni.Name)
	}

	return (compareSlices(expectedUsernames, actualUsernames))
}
