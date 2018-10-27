#!/bin/bash -e

FILE_NAME="zz_generated.deepcopy.go"
cp kong/${FILE_NAME} /tmp

./hack/update-deepcopy-gen.sh

diff -nr /tmp/${FILE_NAME} kong/${FILE_NAME}
