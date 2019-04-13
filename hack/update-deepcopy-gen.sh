#!/bin/bash -e
VERSION="50b56122"

if [[ ! -d /tmp/code-generator ]];
then
  git clone https://github.com/kubernetes/code-generator.git  /tmp/code-generator
  pushd /tmp/code-generator
  git checkout $VERSION
  popd
fi
/tmp/code-generator/generate-groups.sh \
deepcopy \
github.com/hbagdi/go-kong/kong \
github.com/hbagdi \
go-kong:kong \
--go-header-file hack/header-template.go.tmpl
