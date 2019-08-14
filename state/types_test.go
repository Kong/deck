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

func TestUpstreamEqual(t *testing.T) {
	assert := assert.New(t)

	var u1, u2 Upstream
	u1.ID = kong.String("foo")
	u1.Name = kong.String("bar")

	u2.ID = kong.String("foo")
	u2.Name = kong.String("baz")

	assert.False(u1.Equal(&u2))
	assert.False(u1.EqualWithOpts(&u2, false, false))

	u2.Name = kong.String("bar")
	assert.True(u1.Equal(&u2))
	assert.True(u1.EqualWithOpts(&u2, false, false))

	u1.ID = kong.String("fuu")
	assert.False(u1.EqualWithOpts(&u2, false, false))
	assert.True(u1.EqualWithOpts(&u2, true, false))

	var timestamp int64 = 1
	u2.CreatedAt = &timestamp
	assert.False(u1.EqualWithOpts(&u2, false, false))
	assert.False(u1.EqualWithOpts(&u2, false, true))
}

func TestTargetEqual(t *testing.T) {
	assert := assert.New(t)

	var t1, t2 Target
	t1.ID = kong.String("foo")
	t1.Target.Target = kong.String("bar")

	t2.ID = kong.String("foo")
	t2.Target.Target = kong.String("baz")

	assert.False(t1.Equal(&t2))
	assert.False(t1.EqualWithOpts(&t2, false, false, false))

	t2.Target.Target = kong.String("bar")
	assert.True(t1.Equal(&t2))
	assert.True(t1.EqualWithOpts(&t2, false, false, false))

	t1.ID = kong.String("fuu")
	assert.False(t1.EqualWithOpts(&t2, false, false, false))
	assert.True(t1.EqualWithOpts(&t2, true, false, false))

	var timestamp float64 = 1
	t2.CreatedAt = &timestamp
	assert.False(t1.EqualWithOpts(&t2, false, false, false))
	assert.False(t1.EqualWithOpts(&t2, false, true, false))

	t1.Upstream = &kong.Upstream{ID: kong.String("1")}
	t2.Upstream = &kong.Upstream{ID: kong.String("2")}
	assert.False(t1.EqualWithOpts(&t2, true, true, false))
	assert.True(t1.EqualWithOpts(&t2, true, true, true))

	t1.Upstream = &kong.Upstream{ID: kong.String("2")}
	assert.True(t1.EqualWithOpts(&t2, true, true, false))
}

func TestCertificateEqual(t *testing.T) {
	assert := assert.New(t)

	var c1, c2 Certificate
	c1.ID = kong.String("foo")
	c1.Cert = kong.String("certfoo")
	c1.Key = kong.String("keyfoo")

	c2.ID = kong.String("foo")
	c2.Cert = kong.String("certfoo")
	c2.Key = kong.String("keyfoo-unequal")

	assert.False(c1.Equal(&c2))
	assert.False(c1.EqualWithOpts(&c2, false, false))

	c2.Key = kong.String("keyfoo")
	assert.True(c1.Equal(&c2))
	assert.True(c1.EqualWithOpts(&c2, false, false))

	c1.ID = kong.String("fuu")
	assert.False(c1.EqualWithOpts(&c2, false, false))
	assert.True(c1.EqualWithOpts(&c2, true, false))

	var timestamp int64 = 1
	c2.CreatedAt = &timestamp
	assert.False(c1.EqualWithOpts(&c2, false, false))
	assert.False(c1.EqualWithOpts(&c2, false, true))
}

func TestSNIEqual(t *testing.T) {
	assert := assert.New(t)

	var s1, s2 SNI
	s1.ID = kong.String("foo")
	s1.Name = kong.String("bar")

	s2.ID = kong.String("foo")
	s2.Name = kong.String("baz")

	assert.False(s1.Equal(&s2))
	assert.False(s1.EqualWithOpts(&s2, false, false, false))

	s2.Name = kong.String("bar")
	assert.True(s1.Equal(&s2))
	assert.True(s1.EqualWithOpts(&s2, false, false, false))

	s1.ID = kong.String("fuu")
	assert.False(s1.EqualWithOpts(&s2, false, false, false))
	assert.True(s1.EqualWithOpts(&s2, true, false, false))

	var timestamp int64 = 1
	s2.CreatedAt = &timestamp
	assert.False(s1.EqualWithOpts(&s2, false, false, false))
	assert.False(s1.EqualWithOpts(&s2, false, true, false))

	s1.Certificate = &kong.Certificate{ID: kong.String("1")}
	s2.Certificate = &kong.Certificate{ID: kong.String("2")}
	assert.False(s1.EqualWithOpts(&s2, true, true, false))
	assert.True(s1.EqualWithOpts(&s2, true, true, true))

	s1.Certificate = &kong.Certificate{ID: kong.String("2")}
	assert.True(s1.EqualWithOpts(&s2, true, true, false))
}

func TestPluginEqual(t *testing.T) {
	assert := assert.New(t)

	var p1, p2 Plugin
	p1.ID = kong.String("foo")
	p1.Name = kong.String("bar")

	p2.ID = kong.String("foo")
	p2.Name = kong.String("baz")

	assert.False(p1.Equal(&p2))
	assert.False(p1.EqualWithOpts(&p2, false, false, false))

	p2.Name = kong.String("bar")
	assert.True(p1.Equal(&p2))
	assert.True(p1.EqualWithOpts(&p2, false, false, false))

	p1.ID = kong.String("fuu")
	assert.False(p1.EqualWithOpts(&p2, false, false, false))
	assert.True(p1.EqualWithOpts(&p2, true, false, false))

	timestamp := 1
	p2.CreatedAt = &timestamp
	assert.False(p1.EqualWithOpts(&p2, false, false, false))
	assert.False(p1.EqualWithOpts(&p2, false, true, false))

	p1.Service = &kong.Service{ID: kong.String("1")}
	p2.Service = &kong.Service{ID: kong.String("2")}
	assert.False(p1.EqualWithOpts(&p2, true, true, false))
	assert.True(p1.EqualWithOpts(&p2, true, true, true))

	p1.Service = &kong.Service{ID: kong.String("2")}
	assert.True(p1.EqualWithOpts(&p2, true, true, false))
}

func TestConsumerEqual(t *testing.T) {
	assert := assert.New(t)

	var c1, c2 Consumer
	c1.ID = kong.String("foo")
	c1.Username = kong.String("bar")

	c2.ID = kong.String("foo")
	c2.Username = kong.String("baz")

	assert.False(c1.Equal(&c2))
	assert.False(c1.EqualWithOpts(&c2, false, false))

	c2.Username = kong.String("bar")
	assert.True(c1.Equal(&c2))
	assert.True(c1.EqualWithOpts(&c2, false, false))

	c1.ID = kong.String("fuu")
	assert.False(c1.EqualWithOpts(&c2, false, false))
	assert.True(c1.EqualWithOpts(&c2, true, false))

	var a int64 = 1
	c2.CreatedAt = &a
	assert.False(c1.EqualWithOpts(&c2, false, false))
	assert.False(c1.EqualWithOpts(&c2, false, true))
}

func TestKeyAuthEqual(t *testing.T) {
	assert := assert.New(t)

	var k1, k2 KeyAuth
	k1.ID = kong.String("foo")
	k1.Key = kong.String("bar")

	k2.ID = kong.String("foo")
	k2.Key = kong.String("baz")

	assert.False(k1.Equal(&k2))
	assert.False(k1.EqualWithOpts(&k2, false, false, false))

	k2.Key = kong.String("bar")
	assert.True(k1.Equal(&k2))
	assert.True(k1.EqualWithOpts(&k2, false, false, false))

	k1.ID = kong.String("fuu")
	assert.False(k1.EqualWithOpts(&k2, false, false, false))
	assert.True(k1.EqualWithOpts(&k2, true, false, false))

	k2.CreatedAt = kong.Int(1)
	assert.False(k1.EqualWithOpts(&k2, false, false, false))
	assert.False(k1.EqualWithOpts(&k2, false, true, false))

	k2.Consumer = &kong.Consumer{Username: kong.String("u1")}
	assert.False(k1.EqualWithOpts(&k2, false, true, false))
	assert.False(k1.EqualWithOpts(&k2, false, false, true))
}

func TestHMACAuthEqual(t *testing.T) {
	assert := assert.New(t)

	var k1, k2 HMACAuth
	k1.ID = kong.String("foo")
	k1.Username = kong.String("bar")

	k2.ID = kong.String("foo")
	k2.Username = kong.String("baz")

	assert.False(k1.Equal(&k2))
	assert.False(k1.EqualWithOpts(&k2, false, false, false))

	k2.Username = kong.String("bar")
	assert.True(k1.Equal(&k2))
	assert.True(k1.EqualWithOpts(&k2, false, false, false))

	k1.ID = kong.String("fuu")
	assert.False(k1.EqualWithOpts(&k2, false, false, false))
	assert.True(k1.EqualWithOpts(&k2, true, false, false))

	k2.CreatedAt = kong.Int(1)
	assert.False(k1.EqualWithOpts(&k2, false, false, false))
	assert.False(k1.EqualWithOpts(&k2, false, true, false))

	k2.Consumer = &kong.Consumer{Username: kong.String("u1")}
	assert.False(k1.EqualWithOpts(&k2, false, true, false))
	assert.False(k1.EqualWithOpts(&k2, false, false, true))
}

func TestJWTAuthEqual(t *testing.T) {
	assert := assert.New(t)

	var k1, k2 JWTAuth
	k1.ID = kong.String("foo")
	k1.Key = kong.String("bar")

	k2.ID = kong.String("foo")
	k2.Key = kong.String("baz")

	assert.False(k1.Equal(&k2))
	assert.False(k1.EqualWithOpts(&k2, false, false, false))

	k2.Key = kong.String("bar")
	assert.True(k1.Equal(&k2))
	assert.True(k1.EqualWithOpts(&k2, false, false, false))

	k1.ID = kong.String("fuu")
	assert.False(k1.EqualWithOpts(&k2, false, false, false))
	assert.True(k1.EqualWithOpts(&k2, true, false, false))

	k2.CreatedAt = kong.Int(1)
	assert.False(k1.EqualWithOpts(&k2, false, false, false))
	assert.False(k1.EqualWithOpts(&k2, false, true, false))

	k2.Consumer = &kong.Consumer{Username: kong.String("u1")}
	assert.False(k1.EqualWithOpts(&k2, false, true, false))
	assert.False(k1.EqualWithOpts(&k2, false, false, true))
}
