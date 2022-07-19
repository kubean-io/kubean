#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${REPO_ROOT}"/hack/util.sh

HOST_CLUSTER_NAME=${1:-"kubean-host"}

util::delete_cluster "${HOST_CLUSTER_NAME}"
