package state

import (
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func developersCollection() *DevelopersCollection {
	return state().Developers
}

func TestDeveloperInsert(t *testing.T) {
	assert := assert.New(t)
	collection := developersCollection()

	var developer Developer

	assert.NotNil(collection.Add(developer))

	developer.ID = kong.String("first")
	assert.Nil(collection.Add(developer))

	// re-insert
	developer.Email = kong.String("my-name")
	assert.NotNil(collection.Add(developer))
}

func TestDeveloperGetUpdate(t *testing.T) {
	assert := assert.New(t)
	collection := developersCollection()

	var developer Developer
	developer.ID = kong.String("first")
	developer.Email = kong.String("my-name")

	err := collection.Add(developer)
	assert.Nil(err)

	d, err := collection.Get("")
	assert.NotNil(err)
	assert.Nil(d)

	d, err = collection.Get("first")
	assert.Nil(err)
	assert.NotNil(d)

	d.ID = nil
	d.Email = kong.String("my-updated-name")

	err = collection.Update(*d)
	assert.NotNil(err)

	d.ID = kong.String("does-not-exist")
	assert.NotNil(collection.Update(*d))

	d.ID = kong.String("first")
	assert.Nil(collection.Update(*d))

	d, err = collection.Get("my-name")
	assert.NotNil(err)
	assert.Nil(d)

	d, err = collection.Get("my-updated-name")
	assert.Nil(err)
	assert.NotNil(d)
}

// Test to ensure that the memory reference of the pointer returned by Get()
// is different from the one stored in MemDB.
func TestDeveloperGetMemoryReference(t *testing.T) {
	assert := assert.New(t)
	collection := developersCollection()

	var developer Developer
	developer.ID = kong.String("first")
	developer.Email = kong.String("my-name")

	err := collection.Add(developer)
	assert.Nil(err)

	d, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(d)
	d.Email = kong.String("update-should-not-reflect")

	d, err = collection.Get("first")
	assert.Nil(err)
	assert.Equal("my-name", *d.Email)
}

func TestDevelopersInvalidType(t *testing.T) {
	assert := assert.New(t)
	collection := developersCollection()

	type d2 Developer
	var d d2
	d.Email = kong.String("my-name")
	d.ID = kong.String("first")
	txn := collection.db.Txn(true)
	assert.Nil(txn.Insert(developerTableName, &d))
	txn.Commit()

	assert.Panics(func() {
		collection.Get("my-name")
	})
	assert.Panics(func() {
		collection.GetAll()
	})
}

func TestDeveloperDelete(t *testing.T) {
	assert := assert.New(t)
	collection := developersCollection()

	var developer Developer
	developer.ID = kong.String("first")
	developer.Email = kong.String("my-developer")
	err := collection.Add(developer)
	assert.Nil(err)

	d, err := collection.Get("my-developer")
	assert.Nil(err)
	assert.NotNil(d)
	assert.Equal("first", *d.ID)

	err = collection.Delete("first")
	assert.Nil(err)

	err = collection.Delete("")
	assert.NotNil(err)

	err = collection.Delete(*d.ID)
	assert.NotNil(err)
}

func TestDeveloperGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := developersCollection()

	developers := []Developer{
		{
			Developer: kong.Developer{
				ID:    kong.String("first"),
				Email: kong.String("my-developer1"),
			},
		},
		{
			Developer: kong.Developer{
				ID:    kong.String("second"),
				Email: kong.String("my-developer2"),
			},
		},
	}
	for _, s := range developers {
		assert.Nil(collection.Add(s))
	}

	allDevelopers, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(len(developers), len(allDevelopers))
}
