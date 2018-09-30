package custom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRegistry(t *testing.T) {
	assert := assert.New(t)
	r := NewDefaultRegistry()

	assert.NotNil(r)
	var typ Type = "foo"
	entitiy := EntityCRUDDefinition{
		Name: typ,
	}
	err := r.Register(typ, &entitiy)
	assert.Nil(err)
	err = r.Register(typ, &entitiy)
	assert.NotNil(err)

	e := r.Lookup(typ)
	assert.NotNil(e)
	assert.Equal(e, &entitiy)
	e = r.Lookup("NotExists")
	assert.Nil(e)

	err = r.Unregister("NotExists)")
	assert.NotNil(err)

	err = r.Unregister(typ)
	assert.Nil(err)
}
