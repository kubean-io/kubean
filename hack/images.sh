#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
IMAGE_MANIFEST_PATH="${REPO_ROOT}/hack/images.manifest"
DEFAULT_REGISTRY="ghcr.io/kubean-io"
REPO_PATTERN='.*/.*'

# Usage: images::manifest "v0.3.18-19-g5ca4e155" "" "release.daocloud.io/kpanda"
# Return: All images with repo (if not specified in hack/images.manifest will be the default)
#         and tags are required in installing kpanda
function images::manifest() {
    local kubean_version=$1    # kubean version was ought to be provided
    local default_registry=${2:-$DEFAULT_REGISTRY}

    manifest=$(cat $IMAGE_MANIFEST_PATH | sed "s/#KUBEAN_VERSION#/${kubean_version}/g")

    while IFS=' ' read repo tag; do
        if [[ ! $repo =~ $REPO_PATTERN ]]; then
            repo="${default_registry}/${repo}"
        fi
        echo "${repo}:${tag}"

    done <<< "$manifest"
}
