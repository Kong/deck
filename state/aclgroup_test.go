package state

import (
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func aclGroupsCollection() *ACLGroupsCollection {
	return state().ACLGroups
}

func TestACLGroupInsert(t *testing.T) {
	assert := assert.New(t)
	collection := aclGroupsCollection()

	var aclGroup ACLGroup
	assert.NotNil(collection.Add(aclGroup))

	aclGroup.Group = kong.String("my-group")
	aclGroup.ID = kong.String("first")
	err := collection.Add(aclGroup)
	assert.NotNil(err)

	var aclGroup2 ACLGroup
	aclGroup2.Group = kong.String("my-group")
	aclGroup2.ID = kong.String("first")
	aclGroup2.Consumer = &kong.Consumer{
		ID: kong.String("consumer-id"),
	}
	err = collection.Add(aclGroup2)
	assert.Nil(err)

	// re-insert
	err = collection.Add(aclGroup2)
	assert.NotNil(err)

	// re-insert with a different ID
	aclGroup2.ID = kong.String("second")
	err = collection.Add(aclGroup2)
	assert.NotNil(err)

	// re-insert for different consumer
	aclGroup2.Consumer = &kong.Consumer{
		ID: kong.String("consumer2-id"),
	}
	err = collection.Add(aclGroup2)
	assert.Nil(err)
}

func TestACLGroupGetByID(t *testing.T) {
	assert := assert.New(t)
	collection := aclGroupsCollection()

	var aclGroup ACLGroup
	aclGroup.Group = kong.String("my-group")
	aclGroup.ID = kong.String("first")
	aclGroup.Consumer = &kong.Consumer{
		ID: kong.String("consumer1-id"),
	}

	err := collection.Add(aclGroup)
	assert.Nil(err)

	res, err := collection.GetByID("first")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("my-group", *res.Group)

	res, err = collection.GetByID("my-group")
	assert.NotNil(err)
	assert.Nil(res)

	res, err = collection.GetByID("does-not-exist")
	assert.NotNil(err)
	assert.Nil(res)
}

func TestACLGroupGet(t *testing.T) {
	assert := assert.New(t)
	collection := aclGroupsCollection()

	populateWithACLGroupFixtures(assert, collection)

	res, err := collection.Get("first", "does-not-exist")
	assert.NotNil(err)
	assert.Nil(res)

	res, err = collection.Get("does-not-exist", "my-group12")
	assert.NotNil(err)
	assert.Nil(res)

	res, err = collection.Get("consumer1-id", "my-group12")
	assert.Nil(err)
	assert.NotNil(res)
}

func TestACLGroupUpdate(t *testing.T) {
	assert := assert.New(t)
	collection := aclGroupsCollection()

	var aclGroup ACLGroup
	aclGroup.Group = kong.String("my-group")
	aclGroup.ID = kong.String("first")
	aclGroup.Consumer = &kong.Consumer{
		ID: kong.String("consumer1-id"),
	}

	err := collection.Add(aclGroup)
	assert.Nil(err)

	res, err := collection.Get("consumer1-id", "first")
	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal("my-group", *res.Group)

	res.Group = kong.String("my-group2")
	err = collection.Update(*res)
	assert.Nil(err)

	res, err = collection.Get("consumer1-id", "my-group")
	assert.NotNil(err)
	assert.Nil(res)

	res, err = collection.Get("consumer1-id", "my-group2")
	assert.Nil(err)
	assert.Equal("first", *res.ID)
}

func TestACLGroupDelete(t *testing.T) {
	assert := assert.New(t)
	collection := aclGroupsCollection()

	var aclGroup ACLGroup
	aclGroup.Group = kong.String("my-group1")
	aclGroup.ID = kong.String("first")
	aclGroup.Consumer = &kong.Consumer{
		ID: kong.String("consumer1-id"),
	}
	err := collection.Add(aclGroup)
	assert.Nil(err)

	res, err := collection.Get("consumer1-id", "my-group1")
	assert.Nil(err)
	assert.NotNil(res)

	err = collection.Delete(*res.ID)
	assert.Nil(err)

	res, err = collection.Get("consumer1-id", "my-group1")
	assert.NotNil(err)
	assert.Nil(res)

	// delete a non-existing one
	err = collection.Delete("first")
	assert.NotNil(err)

	err = collection.Delete("my-group1")
	assert.NotNil(err)
}

func TestACLGroupGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := aclGroupsCollection()

	populateWithACLGroupFixtures(assert, collection)

	aclGroups, err := collection.GetAll()
	assert.Nil(err)
	assert.Equal(5, len(aclGroups))
}

func TestACLGroupGetByConsumer(t *testing.T) {
	assert := assert.New(t)
	collection := aclGroupsCollection()

	populateWithACLGroupFixtures(assert, collection)

	aclGroups, err := collection.GetAllByConsumerID("consumer1-id")
	assert.Nil(err)
	assert.Equal(3, len(aclGroups))
}

func populateWithACLGroupFixtures(assert *assert.Assertions,
	collection *ACLGroupsCollection,
) {
	aclGroups := []ACLGroup{
		{
			ACLGroup: kong.ACLGroup{
				Group: kong.String("my-group11"),
				ID:    kong.String("first"),
				Consumer: &kong.Consumer{
					ID: kong.String("consumer1-id"),
				},
			},
		},
		{
			ACLGroup: kong.ACLGroup{
				Group: kong.String("my-group12"),
				ID:    kong.String("second"),
				Consumer: &kong.Consumer{
					ID: kong.String("consumer1-id"),
				},
			},
		},
		{
			ACLGroup: kong.ACLGroup{
				Group: kong.String("my-group13"),
				ID:    kong.String("third"),
				Consumer: &kong.Consumer{
					ID: kong.String("consumer1-id"),
				},
			},
		},
		{
			ACLGroup: kong.ACLGroup{
				Group: kong.String("my-group21"),
				ID:    kong.String("fourth"),
				Consumer: &kong.Consumer{
					ID: kong.String("consumer2-id"),
				},
			},
		},
		{
			ACLGroup: kong.ACLGroup{
				Group: kong.String("my-group22"),
				ID:    kong.String("fifth"),
				Consumer: &kong.Consumer{
					ID: kong.String("consumer2-id"),
				},
			},
		},
	}

	for _, k := range aclGroups {
		err := collection.Add(k)
		assert.Nil(err)
	}
}
