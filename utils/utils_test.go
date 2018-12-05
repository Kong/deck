package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmpty(t *testing.T) {
	assert := assert.New(t)
	notEmpty := "not-empty"
	emptyString := ""
	var nilPointer *string
	assert.False(Empty(&notEmpty))
	assert.True(Empty(nilPointer))
	assert.True(Empty(&emptyString))
}

func TestUUID(t *testing.T) {
	assert := assert.New(t)
	uuid := UUID()
	assert.NotEmpty(uuid)
	assert.Regexp(regexp.MustCompile(
		"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"),
		uuid)
}
