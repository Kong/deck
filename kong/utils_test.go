package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringArrayToString(t *testing.T) {
	assert := assert.New(t)

	arr := StringSlice("foo", "bar")
	s := stringArrayToString(arr)
	assert.Equal("[ foo, bar ]", s)

	arr = StringSlice("foo")
	s = stringArrayToString(arr)
	assert.Equal("[ foo ]", s)

	assert.Equal(stringArrayToString(nil), "nil")
}

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

	arrp := StringSlice()
	assert.Empty(arrp)

	arrp = StringSlice("foo", "bar")
	assert.Equal(2, len(arrp))
	assert.Equal("foo", *arrp[0])
	assert.Equal("bar", *arrp[1])
}
