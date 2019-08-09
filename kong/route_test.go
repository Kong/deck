package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteString(T *testing.T) {
	assert := assert.New(T)

	r := &Route{
		Name:         String("foo"),
		Methods:      StringSlice("GET", "POST"),
		Paths:        StringSlice("/foo", "/bar"),
		Hosts:        StringSlice("host1.com", "host2.com"),
		StripPath:    Bool(false),
		PreserveHost: Bool(true),
		SNIs:         StringSlice("snihost1.com", "snihost2.com"),
		Sources: []*CIDRPort{
			{
				IP:   String("10.0.0.0/8"),
				Port: Int(80),
			},
			{
				IP:   String("1.1.0.0/16"),
				Port: Int(80),
			},
		},
		Destinations: []*CIDRPort{
			{
				IP:   String("172.16.0.0/12"),
				Port: Int(443),
			},
			{
				IP:   String("192.168.0.0"),
				Port: Int(80),
			},
		},
	}
	assert.Equal("[ nil foo [ GET, POST ] [ host1.com, host2.com ] "+
		"[ /foo, /bar ] [ snihost1.com, snihost2.com ] "+
		"[ [ 10.0.0.0/8 80 ], [ 1.1.0.0/16 80 ] ] "+
		"[ [ 172.16.0.0/12 443 ], [ 192.168.0.0 80 ] ] "+
		"true false nil nil ]", r.String())

	r = &Route{}
	assert.Equal("[ nil nil nil nil nil nil nil nil nil nil nil nil ]",
		r.String())
}
