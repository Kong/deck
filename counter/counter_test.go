package counter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	var c Counter
	assert := assert.New(t)
	assert.Equal(uint64(0), c.Value())
	assert.Equal(uint64(1), c.Inc())
	assert.Equal(uint64(1), c.Value())
	c.Reset()
	assert.Equal(uint64(0), c.Value())
}
