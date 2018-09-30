package custom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntityObject(t *testing.T) {
	assert := assert.New(t)

	var typ Type = "foo"
	var object Object = map[string]interface{}{
		"id":  "unique",
		"key": "value",
	}
	e := NewEntityObject(typ)
	assert.NotNil(e)
	assert.Equal(typ, e.Type())

	e.SetObject(object)
	assert.Equal(object, e.Object())

	e.AddRelation("bar", "baz")
	e.AddRelation("yo", "yoyo")
	assert.Equal("baz", e.GetRelation("bar"))
	assert.Equal("yoyo", e.GetRelation("yo"))

	assert.Equal(2, len(e.GetAllRelations()))

	assert.Equal("", e.GetRelation("none"))
}
