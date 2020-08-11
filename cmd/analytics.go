package cmd

import (
	"context"
	"net/http"
	"os"
	"time"
)

func sendAnalytics() {
	const (
		minOSArgs        = 2
		analyticsTimeout = 3 * time.Second
	)

	if os.Getenv("DECK_ANALYTICS") == "off" {
		return
	}

	if len(os.Args) < minOSArgs {
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

	ctx, cancel := context.WithDeadline(context.Background(),
		time.Now().Add(analyticsTimeout))
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", URL, nil)
	req.Header["deck-version"] = []string{VERSION}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}
