#!/bin/bash -e

go install k8s.io/code-generator/cmd/deepcopy-gen
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# konnect package
deepcopy-gen --input-dirs github.com/kong/go-database-reconciler/pkg/konnect \
  -O zz_generated.deepcopy \
  --go-header-file scripts/header-template.go.tmpl \
  --output-base $TMP_DIR

cp $TMP_DIR/github.com/kong/go-database-reconciler/pkg/konnect/zz_generated.deepcopy.go \
  konnect/zz_generated.deepcopy.go

# file package
deepcopy-gen --input-dirs github.com/kong/go-database-reconciler/pkg/file \
  -O zz_generated.deepcopy \
  --go-header-file scripts/header-template.go.tmpl \
  --output-base $TMP_DIR

cp $TMP_DIR/github.com/kong/go-database-reconciler/pkg/file/zz_generated.deepcopy.go \
  file/zz_generated.deepcopy.go
