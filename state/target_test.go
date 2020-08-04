package state

import (
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func targetsCollection() *TargetsCollection {
	return state().Targets
}

func TestTargetInsert(t *testing.T) {
	assert := assert.New(t)
	collection := targetsCollection()

	var t0 Target
	t0.Target.Target = kong.String("my-target")
	err := collection.Add(t0)
	assert.NotNil(err)

	t0.ID = kong.String("first")
	err = collection.Add(t0)
	assert.NotNil(err)

	var t1 Target
	t1.Target.Target = kong.String("my-target")
	t1.ID = kong.String("first")
	t1.Upstream = &kong.Upstream{
		ID: kong.String("upstream1-id"),
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
		Name: kong.String("upstream1-id"),
	}
	err = collection.Add(t3)
	assert.NotNil(err)
}

func TestTargetGetUpdate(t *testing.T) {
	assert := assert.New(t)
	collection := targetsCollection()

	var target Target
	target.Target.Target = kong.String("my-target")
	target.ID = kong.String("first")
	target.Upstream = &kong.Upstream{
		ID: kong.String("upstream1-id"),
	}
	assert.NotNil(target.Upstream)
	err := collection.Add(target)
	assert.Nil(err)

	re, err := collection.Get("upstream1-id", "first")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Equal("my-target", *re.Target.Target)

	re.ID = nil
	re.Upstream.ID = nil
	assert.NotNil(collection.Update(*re))

	re.ID = kong.String("does-not-exist")
	assert.NotNil(collection.Update(*re))

	re.ID = kong.String("first")
	assert.NotNil(collection.Update(*re))

	re.Upstream.ID = kong.String("upstream1-id")
	assert.Nil(collection.Update(*re))

	re.Upstream.ID = kong.String("upstream2-id")
	assert.Nil(collection.Update(*re))
}

// Regression test
// to ensure that the memory reference of the pointer returned by Get()
// is different from the one stored in MemDB.
func TestTargetGetMemoryReference(t *testing.T) {
	assert := assert.New(t)
	collection := targetsCollection()

	var target Target
	target.Target.Target = kong.String("my-target")
	target.ID = kong.String("first")
	target.Upstream = &kong.Upstream{
		ID: kong.String("upstream1-id"),
	}
	err := collection.Add(target)
	assert.Nil(err)

	re, err := collection.Get("upstream1-id", "first")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Equal("my-target", *re.Target.Target)

	re.Weight = kong.Int(1)

	re, err = collection.Get("upstream1-id", "my-target")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Nil(re.Weight)
}

func TestTargetsInvalidType(t *testing.T) {
	assert := assert.New(t)

	collection := targetsCollection()

	var upstream Upstream
	upstream.Name = kong.String("my-upstream")
	upstream.ID = kong.String("first")
	txn := collection.db.Txn(true)
	err := txn.Insert(targetTableName, &upstream)
	assert.NotNil(err)
	txn.Abort()

	type badTarget struct {
		kong.Target
		Meta
	}

	target := badTarget{
		Target: kong.Target{
			ID:     kong.String("id"),
			Target: kong.String("target"),
			Upstream: &kong.Upstream{
				ID: kong.String("upstream-id"),
			},
		},
	}

	txn = collection.db.Txn(true)
	err = txn.Insert(targetTableName, &target)
	assert.Nil(err)
	txn.Commit()

	assert.Panics(func() {
		collection.Get("upstream-id", "id")
	})

	assert.Panics(func() {
		collection.GetAll()
	})
}

func TestTargetDelete(t *testing.T) {
	assert := assert.New(t)
	collection := targetsCollection()

	var target Target
	target.Target.Target = kong.String("my-target")
	target.ID = kong.String("first")
	target.Upstream = &kong.Upstream{
		ID: kong.String("upstream1-id"),
	}
	err := collection.Add(target)
	assert.Nil(err)

	re, err := collection.Get("upstream1-id", "my-target")
	assert.Nil(err)
	assert.NotNil(re)

	err = collection.Delete("upstream1-id", *re.ID)
	assert.Nil(err)

	err = collection.Delete("upstream1-id", *re.ID)
	assert.NotNil(err)

	err = collection.Delete("", "first")
	assert.NotNil(err)

	err = collection.Delete("foo", "")
	assert.NotNil(err)
}

func TestTargetGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := targetsCollection()

	var target Target
	target.Target.Target = kong.String("my-target1")
	target.ID = kong.String("first")
	target.Upstream = &kong.Upstream{
		ID: kong.String("upstream1-id"),
	}
	err := collection.Add(target)
	assert.Nil(err)

	var target2 Target
	target2.Target.Target = kong.String("my-target2")
	target2.ID = kong.String("second")
	target2.Upstream = &kong.Upstream{
		ID: kong.String("upstream1-id"),
	}
	err = collection.Add(target2)
	assert.Nil(err)

	targets, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(2, len(targets))
}

func TestTargetGetAllByUpstreamName(t *testing.T) {
	assert := assert.New(t)
	collection := targetsCollection()

	targets := []*Target{
		{
			Target: kong.Target{
				ID:     kong.String("target1-id"),
				Target: kong.String("target1-name"),
				Upstream: &kong.Upstream{
					ID: kong.String("upstream1-id"),
				},
			},
		},
		{
			Target: kong.Target{
				ID:     kong.String("target2-id"),
				Target: kong.String("target2-name"),
				Upstream: &kong.Upstream{
					ID: kong.String("upstream1-id"),
				},
			},
		},
		{
			Target: kong.Target{
				ID:     kong.String("target3-id"),
				Target: kong.String("target3-name"),
				Upstream: &kong.Upstream{
					ID: kong.String("upstream2-id"),
				},
			},
		},
		{
			Target: kong.Target{
				ID:     kong.String("target4-id"),
				Target: kong.String("target4-name"),
				Upstream: &kong.Upstream{
					ID: kong.String("upstream2-id"),
				},
			},
		},
	}

	for _, target := range targets {
		err := collection.Add(*target)
		assert.Nil(err)
	}

	targets, err := collection.GetAllByUpstreamID("upstream1-id")
	assert.Nil(err)
	assert.Equal(2, len(targets))
}
