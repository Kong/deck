package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestRBACRoleService(T *testing.T) {
	runWhenEnterprise(T, ">=0.33.0", true)
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	role := &RBACRole{
		Name: String("roleA"),
	}

	createdRole, err := client.RBACRoles.Create(defaultCtx, role)
	assert.Nil(err)
	assert.NotNil(createdRole)

	role, err = client.RBACRoles.Get(defaultCtx, createdRole.ID)
	assert.Nil(err)
	assert.NotNil(role)

	role.Comment = String("new comment")
	role, err = client.RBACRoles.Update(defaultCtx, role)
	assert.Nil(err)
	assert.NotNil(role)
	assert.Equal("roleA", *role.Name)

	err = client.RBACRoles.Delete(defaultCtx, createdRole.ID)
	assert.Nil(err)

	// ID can be specified
	id := uuid.NewV4().String()
	role = &RBACRole{
		Name: String("teamB"),
		ID:   String(id),
	}

	createdRole, err = client.RBACRoles.Create(defaultCtx, role)
	assert.Nil(err)
	assert.NotNil(createdRole)
	assert.Equal(id, *createdRole.ID)

	err = client.RBACRoles.Delete(defaultCtx, createdRole.ID)
	assert.Nil(err)
}
func TestRBACRoleServiceList(T *testing.T) {
	runWhenEnterprise(T, ">=0.33.0", true)
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	roleA := &RBACRole{
		Name: String("roleA"),
	}
	roleB := &RBACRole{
		Name: String("roleB"),
	}

	createdRoleA, err := client.RBACRoles.Create(defaultCtx, roleA)
	assert.Nil(err)
	createdRoleB, err := client.RBACRoles.Create(defaultCtx, roleB)
	assert.Nil(err)

	roles, err := client.RBACRoles.List(defaultCtx)
	assert.Nil(err)
	assert.NotNil(roles)
	// Counts default roles (super-admin, admin, read-only)
	assert.Equal(5, len(roles))

	err = client.RBACRoles.Delete(defaultCtx, createdRoleA.ID)
	assert.Nil(err)
	err = client.RBACRoles.Delete(defaultCtx, createdRoleB.ID)
	assert.Nil(err)
}
