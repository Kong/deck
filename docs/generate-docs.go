package main

import (
	"flag"
	"log"

	"github.com/kong/deck/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	var outputPath string
	flag.StringVar(&outputPath, "output-path", ".", "path to output directory")
	flag.Parse()

	err := doc.GenMarkdownTree(cmd.NewRootCmd(), outputPath)
	if err != nil {
		log.Fatal(err)
	}
}
