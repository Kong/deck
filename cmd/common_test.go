package cmd

import (
	"github.com/kong/deck/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"github.com/blang/semver/v4"
)

func Test_kongVersion(T *testing.T) {
	kongVersion, _ := os.LookupEnv("KONG_VERSION")
	expectedVersion = semver.MustParse(kongVersion)
	version, err := kongVersion(nil, NewTestClientConfig())
	assert := assert.New(T)
	assert.Nil(err)
	assert.NotNil(version)
	assert.Equal(version, expectedVersion, "The two version should be identical")
}

func NewTestClientConfig() utils.KongClientConfig {
	kongAdminToken, _ := os.LookupEnv("KONG_ADMIN_TOKEN")
	return utils.KongClientConfig{
		Address: "http://localhost:8001",
		Headers: []string{"kong-admin-token:" + kongAdminToken},
	}
}
