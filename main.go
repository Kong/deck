package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kong/deck/cmd"
)

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
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
