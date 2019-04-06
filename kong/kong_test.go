package kong

import (
	"os"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
)

func TestKongStatus(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	status, err := client.Status(nil)
	assert.Nil(err)
	assert.NotNil(status)
}

func TestRoot(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	root, err := client.Root(nil)
	assert.Nil(err)
	assert.NotNil(root)
	assert.NotNil(root["version"])
}

func TestMain(m *testing.M) {
	// to test ListAll code for pagination
	pageSize = 1
	os.Exit(m.Run())
}

var currentVersion semver.Version

// runWhenKong skips the current test if the version of Kong doesn't
// fall in the semverRange.
// This helper function can be used in tests to write version specific
// tests for Kong.
func runWhenKong(t *testing.T, semverRange string) {
	if currentVersion.Major == 0 {
		client, err := NewClient(nil, nil)
		if err != nil {
			t.Error(err)
		}
		res, err := client.Root(nil)
		if err != nil {
			t.Error(err)
		}
		v := res["version"].(string)
		currentVersion, err = semver.Parse(v)
		if err != nil {
			t.Error(err)
		}
	}

	r, err := semver.ParseRange(semverRange)
	if err != nil {
		t.Error(err)
	}
	if !r(currentVersion) {
		t.Skip()
	}

}
