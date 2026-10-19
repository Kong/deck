package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kong/deck/cmd"
)

func registerSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer signal.Stop(sigs)
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
