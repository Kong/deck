#!/bin/bash

set -euo pipefail

source ./.ci/lib.sh

cleanup() {
    local -r resource=${1?resource type required}
    shift

    docker "$resource" ls "$@" \
        --filter "label=$DOCKER_LABEL" \
        --quiet \
    | while read -r id; do
        docker "$resource" rm \
            --force \
            "$id"
    done
}

cleanup container --all
cleanup volume
cleanup network
