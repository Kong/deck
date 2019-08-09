package kong

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpstreamString(T *testing.T) {
	assert := assert.New(T)

	upstream := &Upstream{}
	assert.Equal("[ nil nil ]", upstream.String())

	upstream = &Upstream{
		Name: String("host.com"),
	}
	assert.Equal("[ nil host.com ]", upstream.String())
}

func TestUpstreamMarshal(T *testing.T) {
	assert := assert.New(T)

	upstream := &Upstream{
		Name: String("foo"),
	}

	jsonBytes, err := json.Marshal(&upstream)
	assert.Nil(err)
	assert.JSONEq("{\"name\":\"foo\"}", string(jsonBytes))
}
