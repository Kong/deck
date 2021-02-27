# go-kong

Go bindings for Kong's Admin API

[![GoDoc](https://godoc.org/github.com/kong/go-kong?status.svg)](https://godoc.org/github.com/kong/go-kong/kong)
[![Build Status](https://github.com/kong/go-kong/workflows/CI%20Test/badge.svg)](https://github.com/kong/go-kong/actions?query=branch%3Amain+event%3Apush)
[![Go Report Card](https://goreportcard.com/badge/github.com/kong/go-kong)](https://goreportcard.com/report/github.com/kong/go-kong)

## Importing

```shell
go get github.com/kong/go-kong/kong
```

## Compatibility

`go-kong` is compatible with Kong 1.x and 2.x.
Semantic versioning is followed for versioning `go-kong`.

## Generators

Some code in this repo such as `kong/zz_generated.deepcopy.go` is generated
from API types (see `kong/types.go`).

After making a change to an API type you can run the generators with:

```shell
./hack/update-deepcopy-gen.sh
```

## License

go-kong is licensed with Apache License Version 2.0.
Please read the LICENSE file for more details.
