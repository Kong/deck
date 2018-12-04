package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTargetValid(T *testing.T) {

	assert := assert.New(T)

	target := &Target{}
	assert.Equal(false, target.Valid())
	target = &Target{
		Target:   String("host.com"),
		Upstream: &Upstream{Name: String("upstream-id")},
	}
	assert.Equal(true, target.Valid())
}

func TestTargetString(T *testing.T) {
	assert := assert.New(T)

	target := &Target{}
	assert.Equal("[ nil nil nil ]", target.String())

	target = &Target{
		Target: String("42.42.42.42:42"),
	}
	assert.Equal("[ nil 42.42.42.42:42 nil ]", target.String())
}
