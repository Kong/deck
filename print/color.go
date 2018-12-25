package print

import (
	"github.com/fatih/color"
)

var (
	// CreatePrintf is fmt.Printf with red as foreground color.
	CreatePrintf = color.New(color.FgGreen).PrintfFunc()
	// DeletePrintf is fmt.Printf with green as foreground color.
	DeletePrintf = color.New(color.FgRed).PrintfFunc()
	// UpdatePrintf is fmt.Printf with yellow as foreground color.
	UpdatePrintf = color.New(color.FgYellow).PrintfFunc()

	// CreatePrintln is fmt.Println with red as foreground color.
	CreatePrintln = color.New(color.FgGreen).PrintlnFunc()
	// DeletePrintln is fmt.Println with green as foreground color.
	DeletePrintln = color.New(color.FgRed).PrintlnFunc()
	// UpdatePrintln is fmt.Println with yellow as foreground color.
	UpdatePrintln = color.New(color.FgYellow).PrintlnFunc()
)
