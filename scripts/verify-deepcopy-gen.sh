#!/bin/bash -e

go install k8s.io/code-generator/cmd/deepcopy-gen
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# konnect package
deepcopy-gen --input-dirs github.com/kong/deck/konnect \
  -O zz_generated.deepcopy \
  --go-header-file scripts/header-template.go.tmpl \
  --output-base $TMP_DIR

diff -Naur konnect/zz_generated.deepcopy.go \
  $TMP_DIR/github.com/kong/deck/konnect/zz_generated.deepcopy.go

# file package
deepcopy-gen --input-dirs github.com/kong/deck/file \
  -O zz_generated.deepcopy \
  --go-header-file scripts/header-template.go.tmpl \
  --output-base $TMP_DIR

diff -Naur file/zz_generated.deepcopy.go \
  $TMP_DIR/github.com/kong/deck/file/zz_generated.deepcopy.go
