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
		Fields: []Field{
			{
				Struct: "A",
				Sub:    "Bar",
			},
		},
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
	assert.Nil(err)
	assert.Empty(val)

	s = ""
	ok, val, err = in.FromObject(Baz{
		A: &Foo{
			Bar: &s,
		},
	})
	assert.False(ok)
	assert.Nil(err)
	assert.Empty(val)

	val, err = in.FromArgs("yolo")
	assert.Nil(err)
	assert.Equal("yolo\x00", string(val))

	val, err = in.FromArgs(2)
	assert.Nil(val)
	assert.NotNil(err)

	val, err = in.FromArgs("1", "2")
	assert.Equal([]byte("12\x00"), val)
	assert.Nil(err)
}

func TestSubFieldIndexerPointer(t *testing.T) {
	type Foo struct {
		Bar *string
	}

	type Baz struct {
		A *Foo
	}

	in := &SubFieldIndexer{
		Fields: []Field{
			{
				Struct: "A",
				Sub:    "Bar",
			},
		},
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
