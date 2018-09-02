package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPluginValid(T *testing.T) {
	assert := assert.New(T)

	p := &Plugin{}
	assert.False(p.Valid())

	p = &Plugin{
		Name: String("foo"),
	}
	assert.True(p.Valid())

}

func TestPluginString(T *testing.T) {
	assert := assert.New(T)

	p := &Plugin{}
	assert.Equal("[ nil nil nil nil nil nil map[] ]", p.String())

	p = &Plugin{
		Name: String("bar"),
	}
	assert.Equal("[ nil bar nil nil nil nil map[] ]", p.String())

	p = &Plugin{
		Name: String("foo"),
		Config: map[string]interface{}{
			"array": "value",
		},
	}
	assert.Equal("[ nil foo nil nil nil nil map[array:value] ]", p.String())

	p = &Plugin{
		Name: String("foo"),
		Config: map[string]interface{}{
			"array": []string{
				"element1",
				"element2",
			},
		},
	}
	assert.Equal("[ nil foo nil nil nil nil map[array:[element1 element2]] ]", p.String())

	p = &Plugin{
		Name: String("foo"),
		Config: map[string]interface{}{
			"key": map[string]interface{}{
				"subkey": "subvalue",
			},
		},
	}
	assert.Equal("[ nil foo nil nil nil nil map[key:map[subkey:subvalue]] ]", p.String())
}
