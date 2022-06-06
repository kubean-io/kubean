#!/usr/bin/env bash

set -e

CUR_DIR=$(
    cd -- "$(dirname "$0")" >/dev/null 2>&1
    pwd -P
)

jq .$1 ${CUR_DIR}/../versions.json -r
