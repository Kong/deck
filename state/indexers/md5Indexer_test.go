package indexers

import (
	"crypto/md5"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMD5FieldsIndexer(t *testing.T) {
	assert := assert.New(t)

	type Foo struct {
		Bar *string
		Baz *string
	}

	in := &MD5FieldsIndexer{
		Fields: []string{"Bar", "Baz"},
	}
	s1 := "yolo"
	s2 := "oloy"
	b := Foo{
		Bar: &s1,
		Baz: &s2,
	}

	ok, val, err := in.FromObject(b)
	assert.True(ok)
	assert.Nil(err)
	sum := md5.Sum([]byte("yolooloy"))
	assert.Equal(sum[:], val)

	val, err = in.FromArgs("yolo", "oloy")
	assert.Nil(err)
	assert.Equal(sum[:], val)

	ok, val, err = in.FromObject(Foo{})
	assert.False(ok)
	assert.NotNil(err)

	s1 = ""
	s2 = ""
	ok, val, err = in.FromObject(Foo{
		Bar: &s1,
		Baz: &s2,
	})
	assert.False(ok)
	assert.Nil(err)

	val, err = in.FromArgs("")
	assert.NotNil(err)
	assert.Nil(val)

	val, err = in.FromArgs(2)
	assert.NotNil(err)
	assert.Nil(val)
}
