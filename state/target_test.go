package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func TestTargetInsert(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewTargetsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var t0 Target
	t0.Target.Target = kong.String("my-target")
	t0.ID = kong.String("first")
	err = collection.Add(t0)
	assert.NotNil(err)

	var t1 Target
	t1.Target.Target = kong.String("my-target")
	t1.ID = kong.String("first")
	t1.Upstream = &kong.Upstream{
		ID:   kong.String("upstream1-id"),
		Name: kong.String("upstream1-name"),
	}
	err = collection.Add(t1)
	assert.Nil(err)

	var t2 Target
	t2.Target.Target = kong.String("my-target")
	t2.ID = kong.String("second")
	t2.Upstream = &kong.Upstream{
		ID: kong.String("upstream1-id"),
	}
	err = collection.Add(t2)
	assert.NotNil(err)

	var t3 Target
	t3.Target.Target = kong.String("my-target")
	t3.ID = kong.String("third")
	t3.Upstream = &kong.Upstream{
		Name: kong.String("upstream1-name"),
	}
	err = collection.Add(t3)
	assert.NotNil(err)
}

func TestTargetGetUpdate(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewTargetsCollection()
	assert.Nil(err)
	assert.NotNil(collection)
	var target Target
	target.Target.Target = kong.String("my-target")
	target.ID = kong.String("first")
	target.Upstream = &kong.Upstream{
		ID:   kong.String("upstream1-id"),
		Name: kong.String("upstream1-name"),
	}
	assert.NotNil(target.Upstream)
	err = collection.Add(target)
	assert.NotNil(target.Upstream)
	assert.Nil(err)

	re, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Equal("my-target", *re.Target.Target)
	err = collection.Update(*re)
	assert.Nil(err)

	re, err = collection.GetByUpstreamNameAndTarget("upstream1-name", "my-target")
	assert.Nil(err)
	assert.NotNil(re)
}

// Regression test
// to ensure that the memory reference of the pointer returned by Get()
// is different from the one stored in MemDB.
func TestTargetGetMemoryReference(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewTargetsCollection()
	assert.Nil(err)
	assert.NotNil(collection)
	var target Target
	target.Target.Target = kong.String("my-target")
	target.ID = kong.String("first")
	target.Upstream = &kong.Upstream{
		ID:   kong.String("upstream1-id"),
		Name: kong.String("upstream1-name"),
	}
	err = collection.Add(target)
	assert.Nil(err)

	re, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Equal("my-target", *re.Target.Target)

	re.Weight = kong.Int(1)

	re, err = collection.GetByUpstreamNameAndTarget("upstream1-name", "my-target")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Nil(re.Weight)
}

func TestTargetsInvalidType(t *testing.T) {
	assert := assert.New(t)

	collection, err := NewTargetsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var upstream Upstream
	upstream.Name = kong.String("my-upstream")
	upstream.ID = kong.String("first")
	txn := collection.memdb.Txn(true)
	_ = txn.Insert(targetTableName, &upstream)
	txn.Commit()

	assert.Panics(func() {
		_, _ = collection.Get("first")
	})

	type badTarget struct {
		kong.Target
		Meta
	}

	target := badTarget{
		Target: kong.Target{
			ID:     kong.String("id"),
			Target: kong.String("target"),
			Upstream: &kong.Upstream{
				ID:   kong.String("upstream-id"),
				Name: kong.String("upstream-name"),
			},
		},
	}

	txn = collection.memdb.Txn(true)
	err = txn.Insert(targetTableName, &target)
	assert.Nil(err)
	txn.Commit()

	assert.Panics(func() {
		collection.Get("id")
	})
}

func TestTargetDelete(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewTargetsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var target Target
	target.Target.Target = kong.String("my-target")
	target.ID = kong.String("first")
	target.Upstream = &kong.Upstream{
		ID:   kong.String("upstream1-id"),
		Name: kong.String("upstream1-name"),
	}
	err = collection.Add(target)
	assert.Nil(err)

	re, err := collection.GetByUpstreamNameAndTarget("upstream1-name", "my-target")
	assert.Nil(err)
	assert.NotNil(re)

	err = collection.Delete(*re.ID)
	assert.Nil(err)

	err = collection.Delete(*re.ID)
	assert.NotNil(err)
}

func TestTargetGetAll(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewTargetsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	var target Target
	target.Target.Target = kong.String("my-target1")
	target.ID = kong.String("first")
	target.Upstream = &kong.Upstream{
		ID:   kong.String("upstream1-id"),
		Name: kong.String("upstream1-name"),
	}
	err = collection.Add(target)
	assert.Nil(err)

	var target2 Target
	target2.Target.Target = kong.String("my-target2")
	target2.ID = kong.String("second")
	target2.Upstream = &kong.Upstream{
		ID:   kong.String("upstream1-id"),
		Name: kong.String("upstream1-name"),
	}
	err = collection.Add(target2)
	assert.Nil(err)

	targets, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(2, len(targets))
}

func TestTargetGetAllByUpstreamName(t *testing.T) {
	assert := assert.New(t)
	collection, err := NewTargetsCollection()
	assert.Nil(err)
	assert.NotNil(collection)

	targets := []*Target{
		{
			Target: kong.Target{
				ID:     kong.String("target1-id"),
				Target: kong.String("target1-name"),
				Upstream: &kong.Upstream{
					ID:   kong.String("upstream1-id"),
					Name: kong.String("upstream1-name"),
				},
			},
		},
		{
			Target: kong.Target{
				ID:     kong.String("target2-id"),
				Target: kong.String("target2-name"),
				Upstream: &kong.Upstream{
					ID:   kong.String("upstream1-id"),
					Name: kong.String("upstream1-name"),
				},
			},
		},
		{
			Target: kong.Target{
				ID:     kong.String("target3-id"),
				Target: kong.String("target3-name"),
				Upstream: &kong.Upstream{
					ID:   kong.String("upstream2-id"),
					Name: kong.String("upstream2-name"),
				},
			},
		},
		{
			Target: kong.Target{
				ID:     kong.String("target4-id"),
				Target: kong.String("target4-name"),
				Upstream: &kong.Upstream{
					ID:   kong.String("upstream2-id"),
					Name: kong.String("upstream2-name"),
				},
			},
		},
	}

	for _, target := range targets {
		err = collection.Add(*target)
		assert.Nil(err)
	}

	targets, err = collection.GetAllByUpstreamID("upstream1-id")
	assert.Nil(err)
	assert.Equal(2, len(targets))

	targets, err = collection.GetAllByUpstreamName("upstream2-name")
	assert.Nil(err)
	assert.Equal(2, len(targets))

	targets, err = collection.GetAllByUpstreamName("upstream1-id")
	assert.Nil(err)
	assert.Equal(0, len(targets))
}
