package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCACertificateValid(T *testing.T) {

	assert := assert.New(T)

	c := &CACertificate{}
	assert.Equal(true, c.Valid())
}
