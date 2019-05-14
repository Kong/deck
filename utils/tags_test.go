package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeTags(t *testing.T) {
	type Foo struct {
		Tags []*string
	}
	type Bar struct{}

	assert := assert.New(t)

	a := "tag1"
	b := "tag2"
	c := "tag3"

	var f Foo
	err := MergeTags(f, []string{"tag1"})
	assert.NotNil(err)

	var bar Bar
	err = MergeTags(&bar, []string{"tag1"})
	assert.Nil(err)

	f = Foo{Tags: []*string{&a, &b}}
	MergeTags(&f, []string{"tag1", "tag2", "tag3"})
	assert.True(equalArray([]*string{&a, &b, &c}, f.Tags))

	f = Foo{Tags: []*string{}}
	MergeTags(&f, []string{"tag1", "tag2", "tag3"})
	assert.True(equalArray([]*string{&a, &b, &c}, f.Tags))

	f = Foo{Tags: []*string{&a, &b}}
	MergeTags(&f, nil)
	assert.True(equalArray([]*string{&a, &b}, f.Tags))
}

func equalArray(want, have []*string) bool {
	if len(want) != len(have) {
		return false
	}
	for i := 0; i < len(want); i++ {
		if *want[i] != *have[i] {
			return false
		}
	}
	return true
}

func TestRemoveTags(t *testing.T) {
	type Foo struct {
		Tags []*string
	}
	type Bar struct{}

	assert := assert.New(t)

	a := "tag1"
	b := "tag2"

	var f Foo
	err := RemoveTags(f, []string{"tag1"})
	assert.NotNil(err)

	var bar Bar
	err = RemoveTags(&bar, []string{"tag1"})
	assert.Nil(err)

	f = Foo{Tags: []*string{&a, &b}}
	RemoveTags(&f, []string{"tag2", "tag3"})
	assert.True(equalArray([]*string{&a}, f.Tags))

	f = Foo{Tags: []*string{}}
	RemoveTags(&f, []string{"tag1", "tag2", "tag3"})
	assert.True(equalArray([]*string{}, f.Tags))

	f = Foo{Tags: []*string{&a, &b}}
	RemoveTags(&f, nil)
	assert.True(equalArray([]*string{&a, &b}, f.Tags))
}
