package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hbagdi/deck/cmd"
)

var entities = []string{"key-auth", "hmac-auth", "jwt", "oauth2", "acl"}

func registerSignalHandler() {
	sigs := make(chan os.Signal, 1)
	done := make(chan struct{})
	cmd.SetStopCh(done)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println("received", sig, ", terminating...")
		close(done)
	}()
}

func main() {
	registerSignalHandler()
	cmd.Execute()
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
