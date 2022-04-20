package utils

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/host"
)

const (
	reportsHost = "kong-hf.konghq.com"
	reportsPort = 61829
	konnectMode = "konnect"
)

func SendAnalytics(cmd, deckVersion, kongVersion, mode string) error {
	if strings.ToLower(os.Getenv("DECK_ANALYTICS")) == "off" {
		return nil
	}
	if cmd == "" {
		return fmt.Errorf("invalid argument, 'cmd' cannot be empty")
	}

	stats := collectStats(cmd, deckVersion, kongVersion, mode)
	body := formatStats(stats)

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", reportsHost, reportsPort))
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(body))
	if err != nil {
		return err
	}
	return nil
}

func formatStats(stats map[string]string) string {
	var buffer bytes.Buffer
	buffer.WriteString("<14>")
	for k, v := range stats {
		buffer.WriteString(fmt.Sprintf("%s=%s;", k, v))
	}
	return buffer.String()
}

func collectStats(cmd, deckVersion, kongVersion, mode string) map[string]string {
	result := map[string]string{
		"signal": "decK",
		"v":      deckVersion,
		"cmd":    cmd,
		"os":     runtime.GOOS,
		"arch":   runtime.GOARCH,
	}
	if mode == konnectMode {
		result["mode"] = mode
	}
	if kongVersion != "" && mode != konnectMode {
		result["kv"] = kongVersion
	}
	info, err := host.Info()
	if err == nil {
		result["osv"] = info.Platform + " " + info.PlatformVersion
	}
	return result
}
