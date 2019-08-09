package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsumerString(T *testing.T) {
	assert := assert.New(T)

	c := &Consumer{
		Username: String("bar"),
		CustomID: String("foo"),
	}
	assert.Equal("[ nil bar foo ]", c.String())

	c = &Consumer{
		CustomID: String(""),
		Username: String(""),
	}
	assert.Equal("[ nil nil nil ]", c.String())

	c = &Consumer{}
	assert.Equal("[ nil nil nil ]", c.String())

	c = &Consumer{
		Username: String("foo"),
	}
	assert.Equal("[ nil foo nil ]", c.String())

	c = &Consumer{
		CustomID: String("foo"),
	}
	assert.Equal("[ nil nil foo ]", c.String())
}
