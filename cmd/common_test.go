// +build integration

package cmd

import (
	"context"
	"github.com/blang/semver/v4"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	defaultCtx = context.Background()
)

func Test_kongVersion(T *testing.T) {
	kongVersionEnv, _ := os.LookupEnv("KONG_VERSION")
	var expectedVersion = semver.MustParse(kongVersionEnv)
	var config = NewTestClientConfig("")
	version, err := kongVersion(defaultCtx, config)
	assert := assert.New(T)
	assert.Nil(err)
	assert.NotNil(version)
	assert.Equal(version.Major, expectedVersion.Major, "The two version should have the same major")
	assert.Equal(version.Minor, expectedVersion.Minor, "The two version should have the same minor")

	client, err := utils.GetKongClient(config)
	ws := &kong.Workspace{
		Name: kong.String("test"),
	}
	client.Workspaces.Create(defaultCtx, ws)
	config = NewTestClientConfig(*ws.Name)
	workspaceversion, err := kongVersion(defaultCtx, config)
	assert.Nil(err)
	assert.NotNil(workspaceversion)
	assert.Equal(workspaceversion.Major, expectedVersion.Major, "The two version should have the same major")
	assert.Equal(workspaceversion.Minor, expectedVersion.Minor, "The two version should have the same minor")
	client.Workspaces.Delete(defaultCtx, ws.Name)
}

func NewTestClientConfig(workspace string) utils.KongClientConfig {
	kongAdminToken, _ := os.LookupEnv("KONG_ADMIN_TOKEN")
	return utils.KongClientConfig{
		Address:   "http://localhost:8001",
		Workspace: workspace,
		Headers:   []string{"kong-admin-token:" + kongAdminToken},
	}
}
