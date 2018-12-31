package indexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubFieldIndexer(t *testing.T) {
	type Foo struct {
		Bar *string
	}

	type Baz struct {
		A *Foo
	}

	in := &SubFieldIndexer{
		StructField: "A",
		SubField:    "Bar",
	}
	s := "yolo"
	b := Baz{
		A: &Foo{
			Bar: &s,
		},
	}

	ok, val, err := in.FromObject(b)
	assert := assert.New(t)
	assert.True(ok)
	assert.Nil(err)
	assert.Equal("yolo\x00", string(val))

	ok, val, err = in.FromObject(Baz{})
	assert.False(ok)
	assert.NotNil(err)

	s = ""
	ok, val, err = in.FromObject(Baz{
		A: &Foo{
			Bar: &s,
		},
	})
	assert.False(ok)
	assert.NotNil(err)

	val, err = in.FromArgs("yolo")
	assert.Nil(err)
	assert.Equal("yolo\x00", string(val))

	val, err = in.FromArgs(2)
	assert.Nil(val)
	assert.NotNil(err)

	val, err = in.FromArgs("1", "2")
	assert.Nil(val)
	assert.NotNil(err)
}

func TestSubFieldIndexerPointer(t *testing.T) {
	type Foo struct {
		Bar *string
	}

	type Baz struct {
		A *Foo
	}

	in := &SubFieldIndexer{
		StructField: "A",
		SubField:    "Bar",
	}
	s := "yolo"
	b := Baz{
		A: &Foo{
			Bar: &s,
		},
	}

	ok, val, err := in.FromObject(b)
	assert := assert.New(t)
	assert.True(ok)
	assert.Nil(err)
	assert.Equal("yolo\x00", string(val))

	val, err = in.FromArgs("yolo")
	assert.Nil(err)
	assert.Equal("yolo\x00", string(val))

}
