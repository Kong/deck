package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	assert := assert.New(t)

	s := String("foo")
	assert.Equal("foo", *s)
}

func TestBool(t *testing.T) {
	assert := assert.New(t)

	b := Bool(true)
	assert.Equal(true, *b)
}

func TestInt(t *testing.T) {
	assert := assert.New(t)

	i := Int(42)
	assert.Equal(42, *i)
}

func TestStringSlice(t *testing.T) {
	assert := assert.New(t)

	arr := []string{}
	arrp := StringSlice(arr)
	assert.Empty(arrp)

	arr = []string{"foo", "bar"}
	arrp = StringSlice(arr)
	assert.Equal(2, len(arrp))
	assert.Equal("foo", *arrp[0])
	assert.Equal("bar", *arrp[1])
}
