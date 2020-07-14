package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestACLGroupCreate(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	acl, err := client.ACLs.Create(defaultCtx,
		String("foo"), nil)
	assert.NotNil(err)
	assert.Nil(acl)

	acl = &ACLGroup{}
	acl, err = client.ACLs.Create(defaultCtx, String(""),
		acl)
	assert.NotNil(err)
	assert.Nil(acl)

	acl, err = client.ACLs.Create(defaultCtx,
		String("does-not-exist"), acl)
	assert.NotNil(err)
	assert.Nil(acl)

	// consumer for the ACL group
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	acl = &ACLGroup{
		Group: String("my-group"),
	}
	createdACL, err := client.ACLs.Create(defaultCtx, consumer.ID, acl)
	assert.Nil(err)
	assert.NotNil(createdACL)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestACLGroupCreateWithID(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	acl := &ACLGroup{
		ID:    String(uuid),
		Group: String("my-group"),
	}

	// consumer for the ACLGroup
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdACL, err := client.ACLs.Create(defaultCtx, consumer.ID, acl)
	assert.Nil(err)
	assert.NotNil(createdACL)

	assert.Equal(uuid, *createdACL.ID)
	assert.Equal("my-group", *createdACL.Group)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestACLGroupGet(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	acl := &ACLGroup{
		ID:    String(uuid),
		Group: String("my-group"),
	}

	// consumer for the ACLGroup
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdACL, err := client.ACLs.Create(defaultCtx, consumer.ID, acl)
	assert.Nil(err)
	assert.NotNil(createdACL)

	aclGroup, err := client.ACLs.Get(defaultCtx, consumer.ID, acl.ID)
	assert.Nil(err)
	assert.Equal("my-group", *aclGroup.Group)

	aclGroup, err = client.ACLs.Get(defaultCtx, consumer.ID, acl.Group)
	assert.Nil(err)
	assert.Equal("my-group", *aclGroup.Group)

	aclGroup, err = client.ACLs.Get(defaultCtx, consumer.ID,
		String("does-not-exists"))
	assert.Nil(aclGroup)
	assert.NotNil(err)

	aclGroup, err = client.ACLs.Get(defaultCtx, consumer.ID, String(""))
	assert.Nil(aclGroup)
	assert.NotNil(err)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestACLGroupUpdate(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	acl := &ACLGroup{
		ID:    String(uuid),
		Group: String("my-group"),
	}

	// consumer for the ACLGroup
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdACL, err := client.ACLs.Create(defaultCtx, consumer.ID, acl)
	assert.Nil(err)
	assert.NotNil(createdACL)

	aclGroup, err := client.ACLs.Get(defaultCtx, consumer.ID, acl.ID)
	assert.Nil(err)
	assert.Equal("my-group", *aclGroup.Group)

	acl.Group = String("my-new-group")
	updatedACLGroup, err := client.ACLs.Update(defaultCtx, consumer.ID, acl)
	assert.Nil(err)
	assert.NotNil(updatedACLGroup)
	assert.Equal("my-new-group", *updatedACLGroup.Group)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestACLGroupDelete(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	uuid := uuid.NewV4().String()
	acl := &ACLGroup{
		ID:    String(uuid),
		Group: String("my-group"),
	}

	// consumer for the ACLGroup
	consumer := &Consumer{
		Username: String("foo"),
	}

	consumer, err = client.Consumers.Create(defaultCtx, consumer)
	assert.Nil(err)
	assert.NotNil(consumer)

	createdACL, err := client.ACLs.Create(defaultCtx, consumer.ID, acl)
	assert.Nil(err)
	assert.NotNil(createdACL)

	err = client.ACLs.Delete(defaultCtx, consumer.ID, acl.Group)
	assert.Nil(err)

	aclGroup, err := client.ACLs.Get(defaultCtx, consumer.ID, acl.ID)
	assert.NotNil(err)
	assert.Nil(aclGroup)

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer.ID))
}

func TestACLGroupListMethods(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// consumer for the ACLGroup
	consumer1 := &Consumer{
		Username: String("foo"),
	}

	consumer1, err = client.Consumers.Create(defaultCtx, consumer1)
	assert.Nil(err)
	assert.NotNil(consumer1)

	consumer2 := &Consumer{
		Username: String("bar"),
	}

	consumer2, err = client.Consumers.Create(defaultCtx, consumer2)
	assert.Nil(err)
	assert.NotNil(consumer2)

	// fixtures
	aclGroups := []*ACLGroup{
		{
			Group:    String("acl11"),
			Consumer: consumer1,
		},
		{
			Group:    String("acl12"),
			Consumer: consumer1,
		},
		{
			Group:    String("acl21"),
			Consumer: consumer2,
		},
		{
			Group:    String("acl22"),
			Consumer: consumer2,
		},
	}

	// create fixturs
	for i := 0; i < len(aclGroups); i++ {
		acl, err := client.ACLs.Create(defaultCtx,
			aclGroups[i].Consumer.ID, aclGroups[i])
		assert.Nil(err)
		assert.NotNil(acl)
		aclGroups[i] = acl
	}

	aclGroupsFromKong, next, err := client.ACLs.List(defaultCtx, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(aclGroupsFromKong)
	assert.Equal(4, len(aclGroupsFromKong))

	// first page
	page1, next, err := client.ACLs.List(defaultCtx, &ListOpt{Size: 1})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))

	// last page
	next.Size = 3
	page2, next, err := client.ACLs.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(page2)
	assert.Equal(3, len(page2))

	aclGroupsForConsumer, next, err := client.ACLs.ListForConsumer(defaultCtx,
		consumer1.ID, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(aclGroupsForConsumer)
	assert.Equal(2, len(aclGroupsForConsumer))

	aclGroups, err = client.ACLs.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(aclGroups)
	assert.Equal(4, len(aclGroups))

	assert.Nil(client.Consumers.Delete(defaultCtx, consumer1.ID))
	assert.Nil(client.Consumers.Delete(defaultCtx, consumer2.ID))
}
