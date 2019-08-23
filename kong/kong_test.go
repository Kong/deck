package kong

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	assert := assert.New(t)

	client, err := NewClient(String("foo/bar"), nil)
	assert.Nil(client)
	assert.NotNil(err)
}

func TestKongStatus(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	status, err := client.Status(defaultCtx)
	assert.Nil(err)
	assert.NotNil(status)
}

func TestRoot(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	root, err := client.Root(defaultCtx)
	assert.Nil(err)
	assert.NotNil(root)
	assert.NotNil(root["version"])
}

func TestDo(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	req, err := client.NewRequest("GET", "/does-not-exist", nil, nil)
	assert.Nil(err)
	assert.NotNil(req)
	resp, err := client.Do(nil, req, nil)
	assert.Equal(err, err404{})
	assert.NotNil(resp)
	assert.Equal(404, resp.StatusCode)

	req, err = client.NewRequest("POST", "/", nil, nil)
	assert.Nil(err)
	assert.NotNil(req)
	resp, err = client.Do(nil, req, nil)
	assert.NotNil(resp)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)
	assert.Empty(body)
	assert.Equal(405, resp.StatusCode)
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
		res, err := client.Root(defaultCtx)
		if err != nil {
			t.Error(err)
		}
		v := res["version"].(string)
		rcIndex := strings.Index(v, "rc")
		if rcIndex != -1 {
			v = v[:rcIndex]
		}
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
