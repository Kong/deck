package main

import (
	"math/rand"
	"time"

	"github.com/hbagdi/deck/cmd"
)

var entities = []string{"key-auth", "hmac-auth", "jwt", "oauth2", "acl"}

func main() {
	cmd.Execute()
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
