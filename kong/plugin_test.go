package kong

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPluginString(T *testing.T) {
	assert := assert.New(T)

	p := &Plugin{}
	assert.Equal("[ nil nil nil nil nil nil map[] ]", p.String())

	p = &Plugin{
		Name:  String("bar"),
		RunOn: String("all"),
	}
	assert.Equal("[ nil bar nil nil nil all map[] ]", p.String())

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
		RunOn: String("first"),
	}
	assert.Equal("[ nil foo nil nil nil first "+
		"map[array:[element1 element2]] ]", p.String())

	p = &Plugin{
		Name: String("foo"),
		Config: map[string]interface{}{
			"key": map[string]interface{}{
				"subkey": "subvalue",
			},
		},
	}
	assert.Equal("[ nil foo nil nil nil nil "+
		"map[key:map[subkey:subvalue]] ]", p.String())
}

func TestConfigurationDeepCopyInto(T *testing.T) {
	assert := assert.New(T)

	var c Configuration
	byt := []byte(`{"int":42,"float":4.2,"strings":["foo","bar"]}`)
	if err := json.Unmarshal(byt, &c); err != nil {
		panic(err)
	}

	c2 := c.DeepCopy()
	assert.Equal(c, c2)

	// Both are independent now
	c["int"] = 24
	assert.Equal(24, c["int"])
	assert.Equal(float64(42), c2["int"])

	c["strings"] = []string{"fubar"}
	assert.Equal([]string{"fubar"}, c["strings"].([]string))
	assert.Equal([]interface{}{"foo", "bar"}, c2["strings"])
}
