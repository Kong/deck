package custom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	assert := assert.New(t)

	entity := NewEntityObject("key-auth")
	entity.AddRelation("consumer_id", "bob")
	result, err := render("/consumers/${consumer_id}/key-auths", entity)
	assert.Nil(err)
	assert.Equal(result, "/consumers/bob/key-auths")

	result, err = render("/consumers/${random_id}/key-auths", entity)
	assert.NotNil(err)
	assert.Equal(result, "")
}

func TestEntityCRUDDefinition(t *testing.T) {
	assert := assert.New(t)

	var typ Type = "foo"
	e := EntityCRUDDefinition{
		Name:       typ,
		CRUDPath:   "/consumers/${consumer_id}/foo",
		PrimaryKey: "id",
	}

	var object Object = map[string]interface{}{
		"id":   "unique-id",
		"key1": "value1",
		"key2": "value2",
	}
	entity := NewEntityObject(typ)
	entity.SetObject(object)
	entity.AddRelation("consumer_id", "gopher")

	assert.Equal(typ, e.Type())
	url, err := e.GetEndpoint(entity)
	assert.Nil(err)
	assert.Equal("/consumers/gopher/foo/unique-id", url)

	url, err = e.PatchEndpoint(entity)
	assert.Nil(err)
	assert.Equal("/consumers/gopher/foo/unique-id", url)

	url, err = e.DeleteEndpoint(entity)
	assert.Nil(err)
	assert.Equal("/consumers/gopher/foo/unique-id", url)

	url, err = e.PostEndpoint(entity)
	assert.Nil(err)
	assert.Equal("/consumers/gopher/foo", url)

	url, err = e.ListEndpoint(entity)
	assert.Nil(err)
	assert.Equal("/consumers/gopher/foo", url)

	entity = NewEntityObject(typ)
	entity.SetObject(object)
	url, err = e.GetEndpoint(entity)
	assert.NotNil(err)
	assert.Empty(url)

	object = map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	entity = NewEntityObject(typ)
	entity.AddRelation("consumer_id", "gopher")
	entity.SetObject(object)
	url, err = e.GetEndpoint(entity)
	assert.NotNil(err)
	assert.Empty(url)

	object = map[string]interface{}{
		"id":   -1,
		"key1": "value1",
		"key2": "value2",
	}
	entity = NewEntityObject(typ)
	entity.SetObject(object)
	entity.AddRelation("consumer_id", "gopher")
	url, err = e.GetEndpoint(entity)
	assert.NotNil(err)
	assert.Empty(url)
}
