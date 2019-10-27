package cmd

import (
	"context"
	"net/http"
	"os"
	"time"
)

func sendAnalytics() {
	if os.Getenv("DECK_ANALYTICS") == "off" {
		return
	}
	if len(os.Args) < 2 {
		return
	}
	cmd := os.Args[1]
	if cmd == "help" ||
		cmd == "ping" ||
		cmd == "version" {
		return
	}

	// HTTP to avoid latency due to handshake
	URL := "http://d.yolo42.com/" + cmd

	ctx, _ := context.WithDeadline(context.Background(),
		time.Now().Add(3*time.Second))
	req, _ := http.NewRequestWithContext(ctx, "GET", URL, nil)

	http.DefaultClient.Do(req)
}
