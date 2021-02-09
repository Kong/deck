package state

import (
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func allAvailableCertsCollection() *AllAvailableCertificatesCollection {
	return state().AllAvailableCertificates
}

func TestAllAvailableCertificateInsert(t *testing.T) {
	assert := assert.New(t)
	collection := allAvailableCertsCollection()

	var certificate Certificate
	assert.NotNil(collection.Add(certificate))

	certificate.ID = kong.String("first")
	assert.NotNil(collection.Add(certificate))

	certificate.Key = kong.String("firstKey")
	assert.NotNil(collection.Add(certificate))

	certificate.Cert = kong.String("firstCert")
	err := collection.Add(certificate)
	assert.Nil(err)

	// re-insert
	assert.NotNil(collection.Add(certificate))
}
