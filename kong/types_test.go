package kong

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
