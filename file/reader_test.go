package file

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/kong/deck/dump"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func Test_ensureJSON(t *testing.T) {
	type args struct {
		m map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			"empty array is kept as is",
			args{map[string]interface{}{
				"foo": []interface{}{},
			}},
			map[string]interface{}{
				"foo": []interface{}{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ensureJSON(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ensureJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadKongStateFromStdinFailsToParseText(t *testing.T) {
	filenames := []string{"-"}
	assert := assert.New(t)
	assert.Equal("-", filenames[0])

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

	c, err := GetContentFromFiles(filenames)
	assert.NotNil(err)
	assert.Nil(c)
}

func TestTransformNotFalse(t *testing.T) {
	filenames := []string{"-"}
	assert := assert.New(t)

	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString("_transform: false\nservices:\n- host: test.com\n  name: test service\n"); err != nil {
		panic(err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		panic(err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	os.Stdin = tmpfile

	c, err := GetContentFromFiles(filenames)
	if err != nil {
		panic(err)
	}
	parsed, err := Get(c, RenderConfig{}, dump.Config{})
	assert.Equal(err, ErrorTransformFalseNotSupported)
	assert.Nil(parsed)

	parsed, _, err = GetForKonnect(c, RenderConfig{})
	assert.Equal(err, ErrorTransformFalseNotSupported)
	assert.Nil(parsed)
}

func TestReadKongStateFromStdin(t *testing.T) {
	filenames := []string{"-"}
	assert := assert.New(t)
	assert.Equal("-", filenames[0])

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

	c, err := GetContentFromFiles(filenames)
	assert.NotNil(c)
	assert.Nil(err)

	assert.Equal(kong.Service{
		Name: kong.String("test service"),
		Host: kong.String("test.com"),
	},
		c.Services[0].Service)
}
