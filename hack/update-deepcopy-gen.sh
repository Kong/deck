#!/bin/bash -e

./vendor/k8s.io/code-generator/generate-groups.sh \
deepcopy \
github.com/hbagdi/go-kong/kong \
github.com/hbagdi \
go-kong:kong \
--go-header-file hack/header-template.go.tmpl
