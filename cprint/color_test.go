package cprint

import (
	"bytes"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

// captureOutput captures color.Output and returns the recorded output as
// f runs.
// It is not thread-safe.
func captureOutput(f func()) string {
	backupOutput := color.Output
	defer func() {
		color.Output = backupOutput
	}()
	var out bytes.Buffer
	color.Output = &out
	f()
	return out.String()
}

func TestMain(m *testing.M) {
	backup := color.NoColor
	color.NoColor = false
	exitVal := m.Run()
	color.NoColor = backup
	os.Exit(exitVal)
}

func TestPrint(t *testing.T) {
	tests := []struct {
		name          string
		DisableOutput bool
		Run           func()
		Expected      string
	}{
		{
			name:          "println prints colored output",
			DisableOutput: false,
			Run: func() {
				CreatePrintln("foo")
				UpdatePrintln("bar")
				DeletePrintln("fubaz")
			},
			Expected: "\x1b[32mfoo\x1b[0m\n\x1b[33mbar\x1b[0m\n\x1b[31mfubaz\x1b[0m\n",
		},
		{
			name:          "println doesn't output anything when disabled",
			DisableOutput: true,
			Run: func() {
				CreatePrintln("foo")
				UpdatePrintln("bar")
				DeletePrintln("fubaz")
			},
			Expected: "",
		},
		{
			name:          "printf prints colored output",
			DisableOutput: false,
			Run: func() {
				CreatePrintf("%s", "foo")
				UpdatePrintf("%s", "bar")
				DeletePrintf("%s", "fubaz")
			},
			Expected: "\x1b[32mfoo\x1b[0m\x1b[33mbar\x1b[0m\x1b[31mfubaz\x1b[0m",
		},
		{
			name:          "printf doesn't output anything when disabled",
			DisableOutput: true,
			Run: func() {
				CreatePrintln("foo")
				UpdatePrintln("bar")
				DeletePrintln("fubaz")
			},
			Expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DisableOutput = tt.DisableOutput
			defer func() {
				DisableOutput = false
			}()

			output := captureOutput(func() {
				tt.Run()
			})
			assert.Equal(t, tt.Expected, output)
		})
	}
}
