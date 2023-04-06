package state

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kong/go-kong/kong"
)

func caCertsCollection() *CACertificatesCollection {
	return state().CACertificates
}

func TestCACertificateInsert(t *testing.T) {
	assert := assert.New(t)
	collection := caCertsCollection()

	var caCert CACertificate
	assert.NotNil(collection.Add(caCert))
	caCert.ID = kong.String("first")
	assert.NotNil(collection.Add(caCert))
	caCert.Cert = kong.String("firstCert")
	assert.Nil(collection.Add(caCert))

	// re-inesrt
	assert.NotNil(collection.Add(caCert))
}

func TestCACertificateGetUpdate(t *testing.T) {
	assert := assert.New(t)
	collection := caCertsCollection()

	var caCert CACertificate

	assert.NotNil(collection.Update(caCert))

	caCert.Cert = kong.String("firstCert")
	caCert.ID = kong.String("first")
	assert.NotNil(collection.Update(caCert))

	err := collection.Add(caCert)
	assert.Nil(err)

	se, err := collection.Get("")
	assert.NotNil(err)
	assert.Nil(se)

	se, err = collection.Get("firstCert")
	assert.Nil(err)
	assert.NotNil(se)
	se.Cert = kong.String("firstCert-updated")
	err = collection.Update(*se)
	assert.Nil(err)

	se, err = collection.Get("firstCert-updated")
	assert.Nil(err)
	assert.NotNil(se)
	assert.Equal("firstCert-updated", *se.Cert)

	se, err = collection.Get("not-present")
	assert.Equal(ErrNotFound, err)
	assert.Nil(se)
}

func TestCACertInvalidType(t *testing.T) {
	assert := assert.New(t)
	collection := caCertsCollection()

	var cert Certificate
	cert.Cert = kong.String("my-cert")
	cert.ID = kong.String("first")
	txn := collection.db.Txn(true)
	txn.Insert(caCertTableName, &cert)
	txn.Commit()

	assert.Panics(func() {
		collection.Get("my-cert")
	})
	assert.Panics(func() {
		collection.GetAll()
	})
}

func TestCACertificateDelete(t *testing.T) {
	assert := assert.New(t)
	collection := caCertsCollection()

	assert.NotNil(collection.Delete(""))

	var caCert CACertificate
	caCert.ID = kong.String("first")
	caCert.Cert = kong.String("firstCert")
	err := collection.Add(caCert)
	assert.Nil(err)

	se, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(se)
	assert.Equal("firstCert", *se.Cert)

	err = collection.Delete(*se.ID)
	assert.Nil(err)

	err = collection.Delete(*se.ID)
	assert.NotNil(err)

	caCert.ID = kong.String("first")
	caCert.Cert = kong.String("firstCert")
	err = collection.Add(caCert)
	assert.Nil(err)

	se, err = collection.Get("first")
	assert.Nil(err)
	assert.NotNil(se)
	assert.Equal("firstCert", *se.Cert)

	err = collection.Delete(*se.Cert)
	assert.Nil(err)

	err = collection.Delete(*se.ID)
	assert.NotNil(err)
}

func TestCACertificateGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := caCertsCollection()

	var caCert CACertificate
	caCert.ID = kong.String("first")
	caCert.Cert = kong.String("firstCert")
	err := collection.Add(caCert)
	assert.Nil(err)

	var certificate2 CACertificate
	certificate2.ID = kong.String("second")
	certificate2.Cert = kong.String("secondCert")
	err = collection.Add(certificate2)
	assert.Nil(err)

	certificates, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(2, len(certificates))
}
