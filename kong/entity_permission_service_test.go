package kong

import (
	"net/url"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRBACEntityPermissionservice(T *testing.T) {
	runWhenEnterprise(T, ">=0.33.0", true)
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// Create Workspace
	workspace := &Workspace{
		Name: String("entity-test-workspace"),
	}

	createdWorkspace, err := client.Workspaces.Create(defaultCtx, workspace)
	assert.Nil(err)
	assert.NotNil(createdWorkspace)
	// Create new workspace client
	url, err := url.Parse(defaultBaseURL)
	assert.Nil(err)
	url.Path = path.Join(url.Path, *createdWorkspace.Name)
	workspaceClient, err := NewTestClient(String(url.String()), nil)
	assert.Nil(err)
	assert.NotNil(workspaceClient)
	// Use new client in workspace context.
	role := &RBACRole{
		Name: String("test-role-entity-perm"),
	}

	createdRole, err := workspaceClient.RBACRoles.Create(defaultCtx, role)
	assert.Nil(err)
	assert.NotNil(createdRole)

	// Add Entity Permission to Role
	ep := &RBACEntityPermission{
		Role: &RBACRole{
			ID: createdRole.ID,
		},
		EntityID: String("*"),
		Actions: []*string{
			String("create"),
			String("read"),
		},
	}

	createdEntityPermission, err := workspaceClient.RBACEntityPermissions.Create(defaultCtx, ep)
	assert.Nil(err)
	assert.NotNil(createdEntityPermission)

	ep, err = workspaceClient.RBACEntityPermissions.Get(defaultCtx, createdRole.ID, createdEntityPermission.EntityID)
	assert.Nil(err)
	assert.NotNil(ep)

	ep.Comment = String("new comment")
	ep, err = workspaceClient.RBACEntityPermissions.Update(defaultCtx, ep)
	assert.Nil(err)
	assert.NotNil(ep)
	assert.Equal("new comment", *ep.Comment)

	err = workspaceClient.RBACEntityPermissions.Delete(defaultCtx, createdRole.ID, createdEntityPermission.EntityID)
	assert.Nil(err)
	err = workspaceClient.RBACRoles.Delete(defaultCtx, createdRole.ID)
	assert.Nil(err)
	err = client.Workspaces.Delete(defaultCtx, createdWorkspace.ID)
	assert.Nil(err)

}

func TestRBACEntityPermissionserviceList(T *testing.T) {
	runWhenEnterprise(T, ">=0.33.0", true)
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// Create a workspace
	workspace := &Workspace{
		Name: String("entity-test-list-workspace"),
	}

	createdWorkspace, err := client.Workspaces.Create(defaultCtx, workspace)
	assert.Nil(err)
	assert.NotNil(createdWorkspace)

	// Create a role
	role := &RBACRole{
		Name: String("test-role-entity-perm"),
	}
	createdRole, err := client.RBACRoles.Create(defaultCtx, role)
	assert.Nil(err)
	assert.NotNil(createdRole)

	ep1 := &RBACEntityPermission{
		Role: &RBACRole{
			ID: createdRole.ID,
		},
		EntityID: String("*"),
		Actions: []*string{
			String("create"),
			String("read"),
		},
	}
	ep2 := &RBACEntityPermission{
		Role: &RBACRole{
			ID: createdRole.ID,
		},
		EntityID: createdWorkspace.ID,
		Actions: []*string{
			String("update"),
			String("delete"),
		},
	}

	createdEntityPermissionA, err := client.RBACEntityPermissions.Create(defaultCtx, ep1)
	assert.Nil(err)
	createdEntityPermissionB, err := client.RBACEntityPermissions.Create(defaultCtx, ep2)
	assert.Nil(err)

	eps, err := client.RBACEntityPermissions.ListAllForRole(defaultCtx, createdRole.ID)
	assert.Nil(err)
	assert.NotNil(eps)
	// Counts default ep
	assert.Equal(2, len(eps))

	err = client.RBACEntityPermissions.Delete(defaultCtx, createdRole.ID, createdEntityPermissionA.EntityID)
	assert.Nil(err)
	err = client.RBACEntityPermissions.Delete(defaultCtx, createdRole.ID, createdEntityPermissionB.EntityID)
	assert.Nil(err)
	err = client.RBACRoles.Delete(defaultCtx, createdRole.ID)
	assert.Nil(err)
	err = client.Workspaces.Delete(defaultCtx, createdWorkspace.ID)
	assert.Nil(err)
}
