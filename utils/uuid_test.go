package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUID(t *testing.T) {
	assert := assert.New(t)
	uuid := UUID()
	assert.NotEmpty(uuid)
	assert.Regexp(regexp.MustCompile(
		"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"),
		uuid)
}
