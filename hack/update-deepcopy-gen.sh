#!/bin/bash -e

go install k8s.io/code-generator/cmd/deepcopy-gen
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

deepcopy-gen --input-dirs github.com/kong/go-kong/kong \
  -O zz_generated.deepcopy \
  --go-header-file hack/header-template.go.tmpl \
  --output-base $TMP_DIR

cp $TMP_DIR/github.com/kong/go-kong/kong/zz_generated.deepcopy.go \
  kong/zz_generated.deepcopy.go
