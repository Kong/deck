#!/bin/bash -e

cp file/schema.go /tmp/schema.go
go generate ./...

diff -u /tmp/schema.go file/schema.go
