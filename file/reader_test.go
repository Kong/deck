package file

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/kong/deck/dump"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	tmpfile, err := os.CreateTemp("", "example")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content.Bytes())
	require.NoError(t, err)

	_, err = tmpfile.Seek(0, 0)
	require.NoError(t, err)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	os.Stdin = tmpfile

	c, err := GetContentFromFiles(filenames, false)
	assert.NotNil(err)
	assert.Nil(c)
}

func TestTransformNotFalse(t *testing.T) {
	filenames := []string{"-"}
	assert := assert.New(t)

	tmpfile, err := os.CreateTemp("", "example")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString("_transform: false\nservices:\n- host: test.com\n  name: test service\n")
	require.NoError(t, err)

	_, err = tmpfile.Seek(0, 0)
	require.NoError(t, err)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	os.Stdin = tmpfile

	c, err := GetContentFromFiles(filenames, false)
	require.NoError(t, err)

	ctx := context.Background()
	parsed, err := Get(ctx, c, RenderConfig{}, dump.Config{}, nil)
	assert.Equal(err, ErrorTransformFalseNotSupported)
	assert.Nil(parsed)

	parsed, _, err = GetForKonnect(ctx, c, RenderConfig{}, nil)
	assert.Equal(err, ErrorTransformFalseNotSupported)
	assert.Nil(parsed)
}

func TestReadKongStateFromStdin(t *testing.T) {
	filenames := []string{"-"}
	assert := assert.New(t)
	assert.Equal("-", filenames[0])

	var content bytes.Buffer
	content.Write([]byte("services:\n- host: test.com\n  name: test service\n"))

	tmpfile, err := os.CreateTemp("", "example")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content.Bytes())
	require.NoError(t, err)

	_, err = tmpfile.Seek(0, 0)
	require.NoError(t, err)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	os.Stdin = tmpfile

	c, err := GetContentFromFiles(filenames, false)
	assert.NotNil(c)
	assert.Nil(err)

	assert.Equal(kong.Service{
		Name: kong.String("test service"),
		Host: kong.String("test.com"),
	},
		c.Services[0].Service)
}

func TestReadKongStateFromFile(t *testing.T) {
	filenames := []string{"testdata/config.yaml"}
	assert := assert.New(t)
	assert.Equal("testdata/config.yaml", filenames[0])

	c, err := GetContentFromFiles(filenames, false)
	assert.NotNil(c)
	assert.Nil(err)

	t.Run("enabled field for service is read", func(t *testing.T) {
		assert.Equal(kong.Service{
			Name:    kong.String("svc1"),
			Host:    kong.String("mockbin.org"),
			Enabled: kong.Bool(true),
		}, c.Services[0].Service)
	})
}
