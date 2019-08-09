package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceString(T *testing.T) {
	assert := assert.New(T)

	c := &Service{}
	assert.Equal("[ nil nil nil nil nil nil ]", c.String())

	c = &Service{
		Name:     String("foo"),
		Protocol: String("https"),
		Host:     String("upstream"),
		Port:     Int(1234),
		Path:     String("/foo"),
	}
	assert.Equal("[ nil foo https upstream 1234 /foo ]", c.String())
}
