package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func upstreamsCollection() *UpstreamsCollection {
	return state().Upstreams
}

func TestUpstreamInsert(t *testing.T) {
	assert := assert.New(t)
	collection := upstreamsCollection()

	var upstream Upstream
	upstream.ID = kong.String("first")
	err := collection.Add(upstream)
	assert.NotNil(err)

	var upstream2 Upstream
	upstream2.Name = kong.String("my-upstream")
	upstream2.ID = kong.String("first")
	assert.NotNil(upstream2.Upstream)
	err = collection.Add(upstream2)
	assert.NotNil(upstream2.Upstream)
	assert.Nil(err)
}

func TestUpstreamGetUpdate(t *testing.T) {
	assert := assert.New(t)
	collection := upstreamsCollection()

	var upstream Upstream
	upstream.Name = kong.String("my-upstream")
	upstream.ID = kong.String("first")
	err := collection.Add(upstream)
	assert.Nil(err)

	se, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(se)

	se.Name = kong.String("my-updated-upstream")
	err = collection.Update(*se)
	assert.Nil(err)

	se, err = collection.Get("my-updated-upstream")
	assert.Nil(err)
	assert.NotNil(se)

	se, err = collection.Get("my-upstream")
	assert.Equal(ErrNotFound, err)
	assert.Nil(se)
}

// Regression test
// to ensure that the memory reference of the pointer returned by Get()
// is different from the one stored in MemDB.
func TestUpstreamGetMemoryReference(t *testing.T) {
	assert := assert.New(t)
	collection := upstreamsCollection()

	var upstream Upstream
	upstream.Name = kong.String("my-upstream")
	upstream.ID = kong.String("first")
	err := collection.Add(upstream)
	assert.Nil(err)

	se, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(se)
	se.Slots = kong.Int(1)

	se, err = collection.Get("my-upstream")
	assert.Nil(err)
	assert.NotNil(se)
	assert.Nil(se.Slots)
}
func TestUpstreamsInvalidType(t *testing.T) {
	assert := assert.New(t)

	collection := upstreamsCollection()

	var route Route
	route.Name = kong.String("my-route")
	route.ID = kong.String("first")
	txn := collection.db.Txn(true)
	txn.Insert(upstreamTableName, &route)
	txn.Commit()

	assert.Panics(func() {
		collection.Get("my-route")
	})
}

func TestUpstreamDelete(t *testing.T) {
	assert := assert.New(t)
	collection := upstreamsCollection()

	var upstream Upstream
	upstream.Name = kong.String("my-upstream")
	upstream.ID = kong.String("first")
	err := collection.Add(upstream)
	assert.Nil(err)

	se, err := collection.Get("my-upstream")
	assert.Nil(err)
	assert.NotNil(se)

	err = collection.Delete(*se.ID)
	assert.Nil(err)

	_, err = collection.Get("my-upstream")
	assert.Equal(ErrNotFound, err)

	err = collection.Delete(*se.ID)
	assert.NotNil(err)
}

func TestUpstreamGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := upstreamsCollection()

	var upstream Upstream
	upstream.Name = kong.String("my-upstream1")
	upstream.ID = kong.String("first")
	err := collection.Add(upstream)
	assert.Nil(err)

	var upstream2 Upstream
	upstream2.Name = kong.String("my-upstream2")
	upstream2.ID = kong.String("second")
	err = collection.Add(upstream2)
	assert.Nil(err)

	upstreams, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(2, len(upstreams))
}
