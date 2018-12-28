package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func TestMeta(t *testing.T) {
	assert := assert.New(t)

	var m Meta

	m.AddMeta("foo", "bar")
	r := m.GetMeta("foo")
	res, ok := r.(string)
	assert.True(ok)
	assert.Equal("bar", res)
	// assert.Equal(reflect.TypeOf(r).String(), "string")

	s := "string-pointer"
	m.AddMeta("baz", &s)
	r = m.GetMeta("baz")
	res2, ok := r.(*string)
	assert.True(ok)
	assert.Equal("string-pointer", *res2)

	// can retrive a previous value
	r = m.GetMeta("foo")
	res, ok = r.(string)
	assert.True(ok)
	assert.Equal("bar", res)
}

func TestServiceEqual(t *testing.T) {
	assert := assert.New(t)

	var s1, s2 Service
	s1.ID = kong.String("foo")
	s1.Name = kong.String("bar")

	s2.ID = kong.String("foo")
	s2.Name = kong.String("baz")

	assert.False(s1.Equal(&s2))
	assert.False(s1.EqualWithOpts(&s2, false, false))

	s2.Name = kong.String("bar")
	assert.True(s1.Equal(&s2))
	assert.True(s1.EqualWithOpts(&s2, false, false))

	s1.ID = kong.String("fuu")
	assert.False(s1.EqualWithOpts(&s2, false, false))
	assert.True(s1.EqualWithOpts(&s2, true, false))

	s2.CreatedAt = kong.Int(1)
	s1.UpdatedAt = kong.Int(2)
	assert.False(s1.EqualWithOpts(&s2, false, false))
	assert.False(s1.EqualWithOpts(&s2, false, true))
}

func TestRouteEqual(t *testing.T) {
	assert := assert.New(t)

	var r1, r2 Route
	r1.ID = kong.String("foo")
	r1.Name = kong.String("bar")

	r2.ID = kong.String("foo")
	r2.Name = kong.String("baz")

	assert.False(r1.Equal(&r2))
	assert.False(r1.EqualWithOpts(&r2, false, false, false))

	r2.Name = kong.String("bar")
	assert.True(r1.Equal(&r2))
	assert.True(r1.EqualWithOpts(&r2, false, false, false))

	r1.ID = kong.String("fuu")
	assert.False(r1.EqualWithOpts(&r2, false, false, false))
	assert.True(r1.EqualWithOpts(&r2, true, false, false))

	r2.CreatedAt = kong.Int(1)
	r1.UpdatedAt = kong.Int(2)
	assert.False(r1.EqualWithOpts(&r2, false, false, false))
	assert.False(r1.EqualWithOpts(&r2, false, true, false))
	assert.True(r1.EqualWithOpts(&r2, true, true, false))

	r1.Hosts = kong.StringSlice("demo1.example.com", "demo2.example.com")

	// order matters
	r2.Hosts = kong.StringSlice("demo2.example.com", "demo1.example.com")
	assert.False(r1.EqualWithOpts(&r2, true, true, false))

	r2.Hosts = kong.StringSlice("demo1.example.com", "demo2.example.com")
	assert.True(r1.EqualWithOpts(&r2, true, true, false))

	r1.Service = &kong.Service{ID: kong.String("1")}
	r2.Service = &kong.Service{ID: kong.String("2")}
	assert.False(r1.EqualWithOpts(&r2, true, true, false))
	assert.True(r1.EqualWithOpts(&r2, true, true, true))

	r1.Service = &kong.Service{ID: kong.String("2")}
	assert.True(r1.EqualWithOpts(&r2, true, true, false))
}
