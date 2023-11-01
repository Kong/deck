package cprint

import (
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

// captureOutput captures color.Output and returns the recorded output as
// f runs.
// It is not thread-safe.
func captureOutput(f func()) string {
	// Create a buffer to capture stdout
	var stdoutBuffer string

	// Redirect stdout to the buffer
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()
	r, w, _ := os.Pipe()
	os.Stdout = w

	// write the output
	f()

	// Close the write end of the pipe and read stdout into the buffer
	w.Close()
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	stdoutBuffer = string(buf[:n])

	return stdoutBuffer
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
				CreatePrintln(os.Stdout, "foo")
				UpdatePrintln(os.Stdout, "bar")
				DeletePrintln(os.Stdout, "fubaz")
			},
			Expected: "\x1b[32mfoo\n\x1b[0m\x1b[33mbar\n\x1b[0m\x1b[31mfubaz\n\x1b[0m",
		},
		{
			name:          "println doesn't output anything when disabled",
			DisableOutput: true,
			Run: func() {
				CreatePrintln(os.Stdout, "foo")
				UpdatePrintln(os.Stdout, "bar")
				DeletePrintln(os.Stdout, "fubaz")
			},
			Expected: "",
		},
		{
			name:          "printf prints colored output",
			DisableOutput: false,
			Run: func() {
				CreatePrintf(os.Stdout, "%s", "foo")
				UpdatePrintf(os.Stdout, "%s", "bar")
				DeletePrintf(os.Stdout, "%s", "fubaz")
			},
			Expected: "\x1b[32mfoo\x1b[0m\x1b[33mbar\x1b[0m\x1b[31mfubaz\x1b[0m",
		},
		{
			name:          "printf doesn't output anything when disabled",
			DisableOutput: true,
			Run: func() {
				CreatePrintln(os.Stdout, "foo")
				UpdatePrintln(os.Stdout, "bar")
				DeletePrintln(os.Stdout, "fubaz")
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
