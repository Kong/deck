package crud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpString(t *testing.T) {
	assert := assert.New(t)
	op := Op{"foo"}
	var op2 Op
	assert.Equal("foo", op.String())
	assert.Equal("", op2.String())
}
