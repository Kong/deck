package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func TestCertificateInsert(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewCertificatesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var certificate Certificate
	certificate.ID = kong.String("first")
	certificate.Cert = kong.String("firstCert")
	certificate.Key = kong.String("firstKey")
	err = collection.Add(certificate)
	assert.Nil(err)
}

func TestCertificateGetUpdate(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewCertificatesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var certificate Certificate
	certificate.Cert = kong.String("firstCert")
	certificate.Key = kong.String("firstKey")
	certificate.ID = kong.String("first")
	err = collection.Add(certificate)
	assert.Nil(err)

	se, err := collection.GetByCertKey("firstCert", "firstKey")
	assert.Nil(err)
	assert.NotNil(se)
	se.Cert = kong.String("firstCert-updated")
	err = collection.Update(*se)
	assert.Nil(err)

	se, err = collection.GetByCertKey("firstCert-updated", "firstKey")
	assert.Nil(err)
	assert.NotNil(se)
	assert.Equal("firstCert-updated", *se.Cert)

	se, err = collection.GetByCertKey("not-present", "firstKey")
	assert.Equal(ErrNotFound, err)
	assert.Nil(se)
}

// Regression test
// to ensure that the memory reference of the pointer returned by Get()
// is different from the one stored in MemDB.
func TestCertificateGetMemoryReference(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewCertificatesCollection()
	assert.Nil(err)
	assert.NotNil(collection)
	var cert Certificate
	cert.Cert = kong.String("my-cert")
	cert.Key = kong.String("my-key")
	cert.ID = kong.String("first")
	err = collection.Add(cert)
	assert.Nil(err)

	c, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(c)
	c.Cert = kong.String("my-new-cert")

	c, err = collection.Get("first")
	assert.Nil(err)
	assert.NotNil(c)
	assert.Equal("my-cert", *c.Cert)
}

func TestCertificatesInvalidType(t *testing.T) {
	assert := assert.New(t)

	collection, err := NewCertificatesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var upstream Upstream
	upstream.Name = kong.String("my-upstream")
	upstream.ID = kong.String("first")
	txn := collection.memdb.Txn(true)
	err = txn.Insert(certificateTableName, &upstream)
	assert.NotNil(err)
	txn.Abort()

	type badCertificate struct {
		kong.Certificate
		Meta
	}

	certificate := badCertificate{
		Certificate: kong.Certificate{
			ID:   kong.String("id"),
			Cert: kong.String("Cert"),
			Key:  kong.String("Key"),
		},
	}

	txn = collection.memdb.Txn(true)
	err = txn.Insert(certificateTableName, &certificate)
	assert.Nil(err)
	txn.Commit()

	assert.Panics(func() {
		collection.Get("id")
	})

	assert.Panics(func() {
		collection.GetByCertKey("Cert", "Key")
	})
}

func TestCertificateDelete(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewCertificatesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var certificate Certificate
	certificate.ID = kong.String("first")
	certificate.Cert = kong.String("firstCert")
	certificate.Key = kong.String("firstKey")
	err = collection.Add(certificate)
	assert.Nil(err)

	se, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(se)
	assert.Equal("firstCert", *se.Cert)

	err = collection.Delete(*se.ID)
	assert.Nil(err)

	err = collection.Delete(*se.ID)
	assert.NotNil(err)

	certificate.ID = kong.String("first")
	certificate.Cert = kong.String("firstCert")
	certificate.Key = kong.String("firstKey")
	err = collection.Add(certificate)
	assert.Nil(err)

	se, err = collection.Get("first")
	assert.Nil(err)
	assert.NotNil(se)
	assert.Equal("firstCert", *se.Cert)

	err = collection.DeleteByCertKey(*se.Cert, *se.Key)
	assert.Nil(err)

	err = collection.Delete(*se.ID)
	assert.NotNil(err)
}

func TestCertificateGetAll(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewCertificatesCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var certificate Certificate
	certificate.ID = kong.String("first")
	certificate.Cert = kong.String("firstCert")
	certificate.Key = kong.String("firstKey")
	err = collection.Add(certificate)
	assert.Nil(err)

	var certificate2 Certificate
	certificate2.ID = kong.String("second")
	certificate2.Cert = kong.String("secondCert")
	certificate2.Key = kong.String("secondKey")
	err = collection.Add(certificate2)
	assert.Nil(err)

	certificates, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(2, len(certificates))
}
