package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCertificateValid(T *testing.T) {

	assert := assert.New(T)

	c := &Certificate{}
	assert.Equal(true, c.Valid())
}

func TestCertificateString(T *testing.T) {
	assert := assert.New(T)

	c := &Certificate{}
	assert.Equal("[ nil nil ]", c.String())

	c = &Certificate{
		SNIs: StringSlice("foo.com", "bar.com"),
	}
	assert.Equal("[ nil [ foo.com, bar.com ] ]", c.String())
}
