#!/bin/bash -ex

diff -u <(echo -n) <(gofmt -d -s .)
./scripts/verify-codegen.sh
golint -set_exit_status $(go list ./...)
go vet .
go test ./...
