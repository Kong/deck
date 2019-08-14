package dry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripKey(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("hello", stripKey("hello"))
	assert.Equal("yolo", stripKey("yolo"))
	assert.Equal("world", stripKey("hello world"))
}
