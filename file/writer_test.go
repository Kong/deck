package file

import (
	"bytes"
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
func TestWriteKongStateToStdout(t *testing.T) {
	var ks, _ = state.NewKongState()
	var filename = "-"
	assert := assert.New(t)
	assert.Equal("-", filename)
	assert.NotEmpty(t, ks)
	output := captureOutput(func() {
		KongStateToFile(ks, filename)
	})
	assert.Equal("{}\n", output)

	var service state.Service
	service.ID = kong.String("first")
	service.Host = kong.String("example.com")
	service.Name = kong.String("my-service")
	err := ks.Services.Add(service)
	assert.Nil(err)
	output = captureOutput(func() {
		KongStateToFile(ks, filename)
	})
	assert.Equal("services:\n- host: example.com\n  name: my-service\n", output)

}
