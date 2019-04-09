package file

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadKongStateFromStdinFailsToParseText(t *testing.T) {
	var filename = "-"
	assert := assert.New(t)
	assert.Equal("-", filename)

	var content bytes.Buffer
	content.Write([]byte("hunter2\n"))

	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content.Bytes()); err != nil {
		panic(err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		panic(err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	os.Stdin = tmpfile

	ks, err := GetStateFromFile(filename)
	assert.NotNil(err)
	assert.Empty(ks)
}

func TestReadKongStateFromStdin(t *testing.T) {
	var filename = "-"
	assert := assert.New(t)
	assert.Equal("-", filename)

	var content bytes.Buffer
	content.Write([]byte("services:\n- host: test.com\n  name: test service\n"))

	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content.Bytes()); err != nil {
		panic(err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		panic(err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	os.Stdin = tmpfile

	ks, err := GetStateFromFile(filename)
	assert.NotNil(ks)
	assert.Nil(err)

	services, err := ks.Services.GetAll()
	if err != nil {
		panic(err)
	}
	assert.Equal("test.com", *services[0].Host)
	assert.NotEqual("not.the.same.as.test.com", *services[0].Host)
	assert.Equal("test service", *services[0].Name)
	assert.NotEqual("not the same as 'test service'", *services[0].Name)
}
