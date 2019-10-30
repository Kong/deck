package file

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func captureOutput(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stdout := os.Stdout
	stderr := os.Stderr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
	}()
	os.Stdout = writer
	os.Stderr = writer

	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, reader)
		out <- buf.String()
	}()
	wg.Wait()
	f()
	writer.Close()
	return <-out
}

func TestWriteKongStateToStdoutEmptyState(t *testing.T) {
	var ks, _ = state.NewKongState()
	var filename = "-"
	assert := assert.New(t)
	assert.Equal("-", filename)
	assert.NotEmpty(t, ks)
	// YAML
	output := captureOutput(func() {
		KongStateToFile(ks, WriteConfig{
			Workspace:  "foo",
			Filename:   filename,
			FileFormat: YAML,
		})
	})
	assert.Equal("_format_version: \"1.1\"\n_workspace: foo\n", output)
	// JSON
	output = captureOutput(func() {
		KongStateToFile(ks, WriteConfig{
			Workspace:  "foo",
			Filename:   filename,
			FileFormat: JSON,
		})
	})
	expected := `{
  "_format_version": "1.1",
  "_workspace": "foo"
}`
	assert.Equal(expected, output)
}

func TestWriteKongStateToStdoutStateWithOneService(t *testing.T) {
	var ks, _ = state.NewKongState()
	var filename = "-"
	assert := assert.New(t)
	var service state.Service
	service.ID = kong.String("first")
	service.Host = kong.String("example.com")
	service.Name = kong.String("my-service")
	ks.Services.Add(service)
	// YAML
	output := captureOutput(func() {
		KongStateToFile(ks, WriteConfig{
			Filename:   filename,
			FileFormat: YAML,
		})
	})
	expected := fmt.Sprintf("_format_version: \"1.1\"\nservices:\n- host: %s\n  name: %s\n", *service.Host, *service.Name)
	assert.Equal(expected, output)
	// JSON
	output = captureOutput(func() {
		KongStateToFile(ks, WriteConfig{
			Workspace:  "foo",
			Filename:   filename,
			FileFormat: JSON,
		})
	})
	expected = `{
  "_format_version": "1.1",
  "_workspace": "foo",
  "services": [
    {
      "host": "example.com",
      "name": "my-service"
    }
  ]
}`
	assert.Equal(expected, output)
}

func TestWriteKongStateToStdoutStateWithOneServiceOneRoute(t *testing.T) {
	var ks, _ = state.NewKongState()
	var filename = "-"
	assert := assert.New(t)
	var service state.Service
	service.ID = kong.String("first")
	service.Host = kong.String("example.com")
	service.Name = kong.String("my-service")
	ks.Services.Add(service)

	var route state.Route
	route.Name = kong.String("my-route")
	route.ID = kong.String("first")
	route.Hosts = kong.StringSlice("example.com", "demo.example.com")
	route.Service = &kong.Service{
		ID:   kong.String(*service.ID),
		Name: kong.String(*service.Name),
	}

	ks.Routes.Add(route)
	// YAML
	output := captureOutput(func() {
		KongStateToFile(ks, WriteConfig{
			Filename:   filename,
			FileFormat: YAML,
		})
	})
	expected := fmt.Sprintf(`_format_version: "1.1"
services:
- host: %s
  name: %s
  routes:
  - hosts:
    - %s
    - %s
    name: %s
`, *service.Host, *service.Name, *route.Hosts[0], *route.Hosts[1], *route.Name)
	assert.Equal(expected, output)
	// JSON
	output = captureOutput(func() {
		KongStateToFile(ks, WriteConfig{
			Workspace:  "foo",
			Filename:   filename,
			FileFormat: JSON,
		})
	})
	expected = `{
  "_format_version": "1.1",
  "_workspace": "foo",
  "services": [
    {
      "host": "example.com",
      "name": "my-service",
      "routes": [
        {
          "hosts": [
            "example.com",
            "demo.example.com"
          ],
          "name": "my-route"
        }
      ]
    }
  ]
}`
	assert.Equal(expected, output)
}

func Test_addExtToFilename(t *testing.T) {
	type args struct {
		filename string
		format   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				filename: "foo",
				format:   "yolo",
			},
			want: "foo.yolo",
		},
		{
			args: args{
				filename: "foo.json",
				format:   "yolo",
			},
			want: "foo.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addExtToFilename(tt.args.filename, tt.args.format); got != tt.want {
				t.Errorf("addExtToFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
