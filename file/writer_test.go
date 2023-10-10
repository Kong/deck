package file

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
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

func Test_compareOrder(t *testing.T) {
	tests := []struct {
		name      string
		sortable1 sortable
		sortable2 sortable
		expected  bool
	}{
		{
			sortable1: &FService{
				Service: kong.Service{
					Name: kong.String("my-service-1"),
					ID:   kong.String("my-id-1"),
				},
			},
			sortable2: &FService{
				Service: kong.Service{
					Name: kong.String("my-service-2"),
					ID:   kong.String("my-id-2"),
				},
			},
			expected: true,
		},

		{
			sortable1: &FRoute{
				Route: kong.Route{
					Name: kong.String("my-route-1"),
					ID:   kong.String("my-id-1"),
				},
			},
			sortable2: &FRoute{
				Route: kong.Route{
					Name: kong.String("my-route-2"),
					ID:   kong.String("my-id-2"),
				},
			},
			expected: true,
		},

		{
			sortable1: FUpstream{
				Upstream: kong.Upstream{
					Name: kong.String("my-upstream-1"),
					ID:   kong.String("my-id-1"),
				},
			},
			sortable2: FUpstream{
				Upstream: kong.Upstream{
					Name: kong.String("my-upstream-2"),
					ID:   kong.String("my-id-2"),
				},
			},
			expected: true,
		},

		{
			sortable1: FTarget{
				Target: kong.Target{
					Target: kong.String("my-target-1"),
					ID:     kong.String("my-id-1"),
				},
			},
			sortable2: FTarget{
				Target: kong.Target{
					Target: kong.String("my-target-2"),
					ID:     kong.String("my-id-2"),
				},
			},
			expected: true,
		},

		{
			sortable1: FCertificate{
				Cert: kong.String("my-certificate-1"),
				ID:   kong.String("my-id-1"),
			},
			sortable2: FCertificate{
				Cert: kong.String("my-certificate-2"),
				ID:   kong.String("my-id-2"),
			},
			expected: true,
		},

		{
			sortable1: FCACertificate{
				CACertificate: kong.CACertificate{
					Cert: kong.String("my-ca-certificate-1"),
					ID:   kong.String("my-id-1"),
				},
			},
			sortable2: FCACertificate{
				CACertificate: kong.CACertificate{
					Cert: kong.String("my-ca-certificate-2"),
					ID:   kong.String("my-id-2"),
				},
			},
			expected: true,
		},

		{
			sortable1: FPlugin{
				Plugin: kong.Plugin{
					Name: kong.String("my-plugin-1"),
					ID:   kong.String("my-id-1"),
				},
			},
			sortable2: FPlugin{
				Plugin: kong.Plugin{
					Name: kong.String("my-plugin-2"),
					ID:   kong.String("my-id-2"),
				},
			},
			expected: true,
		},

		{
			sortable1: &FConsumer{
				Consumer: kong.Consumer{
					Username: kong.String("my-consumer-1"),
					ID:       kong.String("my-id-2"),
				},
			},
			sortable2: &FConsumer{
				Consumer: kong.Consumer{
					Username: kong.String("my-consumer-2"),
					ID:       kong.String("my-id-2"),
				},
			},
			expected: true,
		},

		{
			sortable1: &FServicePackage{
				Name: kong.String("my-service-package-1"),
				ID:   kong.String("my-id-1"),
			},
			sortable2: &FServicePackage{
				Name: kong.String("my-service-package-2"),
				ID:   kong.String("my-id-2"),
			},
			expected: true,
		},
		{
			sortable1: &FServiceVersion{
				Version: kong.String("my-service-version-1"),
				ID:      kong.String("my-id-1"),
			},
			sortable2: &FServiceVersion{
				Version: kong.String("my-service-version-2"),
				ID:      kong.String("my-id-2"),
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if compareOrder(tt.sortable1, tt.sortable2) != tt.expected {
				t.Errorf("Expected %v, but isn't", tt.expected)
			}
		})
	}
}

func TestWriteKongStateToStdoutEmptyState(t *testing.T) {
	ks, _ := state.NewKongState()
	filename := "-"
	assert := assert.New(t)
	assert.Equal("-", filename)
	assert.NotEmpty(t, ks)
	// YAML
	output := captureOutput(func() {
		KongStateToFile(ks, WriteConfig{
			Workspace:   "foo",
			Filename:    filename,
			FileFormat:  YAML,
			KongVersion: "2.8.0",
		})
	})
	assert.Equal("_format_version: \"1.1\"\n_workspace: foo\n", output)
	// JSON
	output = captureOutput(func() {
		KongStateToFile(ks, WriteConfig{
			Workspace:   "foo",
			Filename:    filename,
			FileFormat:  JSON,
			KongVersion: "2.8.0",
		})
	})
	expected := `{
  "_format_version": "1.1",
  "_workspace": "foo"
}`
	assert.Equal(expected, output)
}

func TestWriteKongStateToStdoutStateWithOneService(t *testing.T) {
	ks, _ := state.NewKongState()
	filename := "-"
	assert := assert.New(t)
	var service state.Service
	service.ID = kong.String("first")
	service.Host = kong.String("example.com")
	service.Name = kong.String("my-service")
	ks.Services.Add(service)
	// YAML
	output := captureOutput(func() {
		KongStateToFile(ks, WriteConfig{
			Filename:    filename,
			FileFormat:  YAML,
			KongVersion: "3.0.0",
		})
	})
	expected := fmt.Sprintf("_format_version: \"3.0\"\nservices:\n- host: %s\n  name: %s\n", *service.Host, *service.Name)
	assert.Equal(expected, output)
	// JSON
	output = captureOutput(func() {
		KongStateToFile(ks, WriteConfig{
			Workspace:   "foo",
			Filename:    filename,
			FileFormat:  JSON,
			KongVersion: "3.0.0",
		})
	})
	expected = `{
  "_format_version": "3.0",
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
	ks, _ := state.NewKongState()
	filename := "-"
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
			Filename:    filename,
			FileFormat:  YAML,
			KongVersion: "2.8.0",
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
			Workspace:   "foo",
			Filename:    filename,
			FileFormat:  JSON,
			KongVersion: "2.8.0",
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

func Test_getFormatVersion(t *testing.T) {
	tests := []struct {
		name        string
		kongVersion string
		expected    string
		expectedErr string
		wantErr     bool
	}{
		{
			name:        "3.0.0 version",
			kongVersion: "3.0.0",
			expected:    "3.0",
		},
		{
			name:        "3.0.0.0 version",
			kongVersion: "3.0.0.0",
			expected:    "3.0",
		},
		{
			name:        "2.8.0 version",
			kongVersion: "2.8.0",
			expected:    "1.1",
		},
		{
			name:        "2.8.0.0 version",
			kongVersion: "2.8.0.0",
			expected:    "1.1",
		},
		{
			name:        "2.8.0.1-enterprise-edition version",
			kongVersion: "2.8.0.1-enterprise-edition",
			expected:    "1.1",
		},
		{
			name:        "unsupported version",
			kongVersion: "test",
			wantErr:     true,
			expectedErr: "parsing Kong version: unknown Kong version",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := getFormatVersion(tt.kongVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, expected error = %v", err, tt.wantErr)
			}
			if tt.expectedErr != "" {
				assert.Equal(t, err.Error(), tt.expectedErr)
			}
			if res != tt.expected {
				t.Errorf("Expected %v, but isn't: %v", tt.expected, res)
			}
		})
	}
}