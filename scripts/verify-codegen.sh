#!/bin/bash -e

FILE="kong_json_schema.json"
cp file/${FILE} /tmp/${FILE}
go generate ./...

diff -u /tmp/${FILE} file/${FILE}
