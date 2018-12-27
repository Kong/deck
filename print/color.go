package print

import (
	"sync"

	"github.com/fatih/color"
)

var (
	// mu is used to synchronize writes from multiple goroutines.
	mu sync.Mutex

	createPrintf = color.New(color.FgGreen).PrintfFunc()
	// CreatePrintf is fmt.Printf with red as foreground color.
	CreatePrintf = func(format string, a ...interface{}) {
		mu.Lock()
		defer mu.Unlock()
		createPrintf(format, a...)
	}
	deletePrintf = color.New(color.FgRed).PrintfFunc()
	// DeletePrintf is fmt.Printf with green as foreground color.
	DeletePrintf = func(format string, a ...interface{}) {
		mu.Lock()
		defer mu.Unlock()
		deletePrintf(format, a...)
	}
	updatePrintf = color.New(color.FgYellow).PrintfFunc()
	// UpdatePrintf is fmt.Printf with yellow as foreground color.
	UpdatePrintf = func(format string, a ...interface{}) {
		mu.Lock()
		defer mu.Unlock()
		updatePrintf(format, a...)
	}
	createPrintln = color.New(color.FgGreen).PrintlnFunc()
	// CreatePrintln is fmt.Println with red as foreground color.
	CreatePrintln = func(a ...interface{}) {
		mu.Lock()
		defer mu.Unlock()
		createPrintln(a...)
	}
	deletePrintln = color.New(color.FgRed).PrintlnFunc()
	// DeletePrintln is fmt.Println with green as foreground color.
	DeletePrintln = func(a ...interface{}) {
		mu.Lock()
		defer mu.Unlock()
		deletePrintln(a...)
	}
	updatePrintln = color.New(color.FgYellow).PrintlnFunc()
	// UpdatePrintln is fmt.Println with yellow as foreground color.
	UpdatePrintln = func(a ...interface{}) {
		mu.Lock()
		defer mu.Unlock()
		updatePrintln(a...)
	}
)
