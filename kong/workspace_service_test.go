// +build enterprise

package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestWorkspaceService(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	workspace := &Workspace{
		Name: String("teamA"),
		Meta: map[string]interface{}{
			"color":     "#814CA6",
			"thumbnail": nil,
		},
	}

	createdWorkspace, err := client.Workspaces.Create(defaultCtx, workspace)
	assert.Nil(err)
	assert.NotNil(createdWorkspace)

	workspace, err = client.Workspaces.Get(defaultCtx, createdWorkspace.ID)
	assert.Nil(err)
	assert.NotNil(workspace)

	workspace.Comment = String("new comment")
	workspace, err = client.Workspaces.Update(defaultCtx, workspace)
	assert.Nil(err)
	assert.NotNil(workspace)
	assert.NotNil(workspace.Config)
	assert.Equal("teamA", *workspace.Name)
	assert.Equal("new comment", *workspace.Comment)
	assert.Equal("#814CA6", workspace.Meta["color"])

	err = client.Workspaces.Delete(defaultCtx, createdWorkspace.ID)
	assert.Nil(err)

	// ID can be specified
	id := uuid.NewV4().String()
	workspace = &Workspace{
		Name: String("teamB"),
		ID:   String(id),
	}

	createdWorkspace, err = client.Workspaces.Create(defaultCtx, workspace)
	assert.Nil(err)
	assert.NotNil(createdWorkspace)
	assert.Equal(id, *createdWorkspace.ID)

	err = client.Workspaces.Delete(defaultCtx, createdWorkspace.ID)
	assert.Nil(err)
}

func TestWorkspaceServiceList(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	workspaceA := &Workspace{
		Name: String("teamA"),
	}
	workspaceB := &Workspace{
		Name: String("teamB"),
	}

	createdWorkspaceA, err := client.Workspaces.Create(defaultCtx, workspaceA)
	assert.Nil(err)
	createdWorkspaceB, err := client.Workspaces.Create(defaultCtx, workspaceB)
	assert.Nil(err)
	// paged List
	page1, next, err := client.Workspaces.List(defaultCtx, &ListOpt{Size: 1})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))
	// nil ListOpt List
	workspaces, next, err := client.Workspaces.List(defaultCtx, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(workspaces)
	// Counts default workspace
	assert.Equal(3, len(workspaces))

	err = client.Workspaces.Delete(defaultCtx, createdWorkspaceA.ID)
	assert.Nil(err)
	err = client.Workspaces.Delete(defaultCtx, createdWorkspaceB.ID)
	assert.Nil(err)
}

func TestWorkspaceServiceListAll(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	workspaceA := &Workspace{
		Name: String("teamA"),
	}
	workspaceB := &Workspace{
		Name: String("teamB"),
	}

	createdWorkspaceA, err := client.Workspaces.Create(defaultCtx, workspaceA)
	assert.Nil(err)
	createdWorkspaceB, err := client.Workspaces.Create(defaultCtx, workspaceB)
	assert.Nil(err)

	workspaces, err := client.Workspaces.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(workspaces)
	// Counts default workspace
	assert.Equal(3, len(workspaces))

	err = client.Workspaces.Delete(defaultCtx, createdWorkspaceA.ID)
	assert.Nil(err)
	err = client.Workspaces.Delete(defaultCtx, createdWorkspaceB.ID)
	assert.Nil(err)
}
