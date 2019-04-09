package file

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
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
		log.SetOutput(os.Stderr)
	}()
	os.Stdout = writer
	os.Stderr = writer

	log.SetOutput(writer)
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
	output := captureOutput(func() {
		KongStateToFile(ks, filename)
	})
	assert.Equal("{}\n", output)

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
	output := captureOutput(func() {
		KongStateToFile(ks, filename)
	})
	fmt.Print(service.Host)
	expected := fmt.Sprintf("services:\n- host: %s\n  name: %s\n", *service.Host, *service.Name)
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

	output := captureOutput(func() {
		KongStateToFile(ks, filename)
	})
	fmt.Print(service.Host)
	expected := fmt.Sprintf("services:\n- host: %s\n  name: %s\n  routes:\n  - hosts:\n    - %s\n    - %s\n    name: %s\n", *service.Host, *service.Name, *route.Hosts[0], *route.Hosts[1], *route.Name)
	assert.Equal(expected, output)
}
