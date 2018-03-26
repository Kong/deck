package main

import (
	"fmt"
	"log"

	"github.com/hbagdi/go-kong/kong"
)

func main() {
	client := kong.NewClient(nil)
	status, err := client.Status()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(status)
}
