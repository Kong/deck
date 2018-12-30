package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPlaceHolder(t *testing.T) {
	assert := assert.New(t)

	var res bool
	var s string

	res = isPlaceHolder(nil)
	assert.False(res)

	s = " "
	res = isPlaceHolder(&s)
	assert.False(res)

	s = " placeholder"
	res = isPlaceHolder(&s)
	assert.False(res)

	s = "prefixplaceholder"
	res = isPlaceHolder(&s)
	assert.False(res)

	s = "placeholder"
	res = isPlaceHolder(&s)
	assert.True(res)

	s = "placeholdersuffix"
	res = isPlaceHolder(&s)
	assert.True(res)
}
