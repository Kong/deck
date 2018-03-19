package main

import "github.com/hbagdi/go-kong/kong"

func main() {
	kong := kong.New(nil)
	kong.Sample.Foo()
}
