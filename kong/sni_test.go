package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSNIValid(T *testing.T) {

	assert := assert.New(T)

	sni := &SNI{}
	assert.Equal(false, sni.Valid())
	sni = &SNI{
		Name: String("host.com"),
	}
	assert.Equal(true, sni.Valid())
}

func TestSNIString(T *testing.T) {
	assert := assert.New(T)

	sni := &SNI{}
	assert.Equal("[ nil nil nil ]", sni.String())

	sni = &SNI{
		Name: String("host.com"),
		Certificate: &Certificate{
			ID: String("foo"),
		},
	}
	assert.Equal("[ nil host.com foo ]", sni.String())
}
