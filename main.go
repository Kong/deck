package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kong/deck/cmd"
)

func registerSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println("received", sig, ", terminating...")
		cancel()
	}()
	return ctx
}

func main() {
	ctx := registerSignalHandler()
	cmd.Execute(ctx)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
