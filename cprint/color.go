package cprint

import (
	"io"
	"sync"

	"github.com/fatih/color"
)

var (
	// mu is used to synchronize writes from multiple goroutines.
	mu sync.Mutex
	// DisableOutput disables all output.
	DisableOutput bool
)

func conditionalPrintf(w io.Writer, fn func(io.Writer, string, ...interface{}), format string, a ...interface{}) {
	if DisableOutput {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	fn(w, format, a...)
}

func conditionalPrintln(w io.Writer, fn func(io.Writer, ...interface{}), a ...interface{}) {
	if DisableOutput {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	fn(w, a...)
}

var (
	createPrintf = color.New(color.FgGreen).FprintfFunc()
	deletePrintf = color.New(color.FgRed).FprintfFunc()
	updatePrintf = color.New(color.FgYellow).FprintfFunc()

	// CreatePrintf is fmt.Printf with red as foreground color.
	CreatePrintf = func(w io.Writer, format string, a ...interface{}) {
		conditionalPrintf(w, createPrintf, format, a...)
	}

	// DeletePrintf is fmt.Printf with green as foreground color.
	DeletePrintf = func(w io.Writer, format string, a ...interface{}) {
		conditionalPrintf(w, deletePrintf, format, a...)
	}

	// UpdatePrintf is fmt.Printf with yellow as foreground color.
	UpdatePrintf = func(w io.Writer, format string, a ...interface{}) {
		conditionalPrintf(w, updatePrintf, format, a...)
	}

	createPrintln = color.New(color.FgGreen).FprintlnFunc()
	deletePrintln = color.New(color.FgRed).FprintlnFunc()
	updatePrintln = color.New(color.FgYellow).FprintlnFunc()
	bluePrintln   = color.New(color.BgBlue).FprintlnFunc()

	// CreatePrintln is fmt.Println with red as foreground color.
	CreatePrintln = func(w io.Writer, a ...interface{}) {
		conditionalPrintln(w, createPrintln, a...)
	}

	// DeletePrintln is fmt.Println with green as foreground color.
	DeletePrintln = func(w io.Writer, a ...interface{}) {
		conditionalPrintln(w, deletePrintln, a...)
	}

	// UpdatePrintln is fmt.Println with yellow as foreground color.
	UpdatePrintln = func(w io.Writer, a ...interface{}) {
		conditionalPrintln(w, updatePrintln, a...)
	}

	BluePrintLn = func(w io.Writer, a ...interface{}) {
		conditionalPrintln(w, bluePrintln, a...)
	}
)
