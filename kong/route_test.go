package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteValid(T *testing.T) {

	assert := assert.New(T)

	r := &Route{}
	assert.Equal(false, r.Valid())

	r = &Route{
		Protocols: StringSlice([]string{"http"}),
	}
	assert.Equal(false, r.Valid())
}

func TestRouteString(T *testing.T) {
	assert := assert.New(T)

	r := &Route{
		Methods:      StringSlice([]string{"GET", "POST"}),
		Paths:        StringSlice([]string{"/foo", "/bar"}),
		Hosts:        StringSlice([]string{"host1.com", "host2.com"}),
		StripPath:    Bool(false),
		PreserveHost: Bool(true),
	}
	assert.Equal("[ nil [ GET, POST ] [ host1.com, host2.com ] [ /foo, /bar ] true false nil nil ]", r.String())

	r = &Route{}
	assert.Equal("[ nil nil nil nil nil nil nil nil ]", r.String())
}
