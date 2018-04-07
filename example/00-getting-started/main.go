package main

import (
	"fmt"
	"log"

	"github.com/hbagdi/go-kong/kong"
)

func main() {
	client, err := kong.NewClient(nil, nil)
	if err != nil {
		log.Fatalln(err)
	}
	status, err := client.Status(nil)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(*status)
}
