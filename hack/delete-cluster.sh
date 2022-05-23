#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${REPO_ROOT}"/hack/util.sh

HOST_CLUSTER_NAME=${1:-"kubean-host"}
MEMBER_CLUSTER_1_NAME=${2:-"member1"}
MEMBER_CLUSTER_2_NAME=${3:-"member2"}

util::delete_cluster "${HOST_CLUSTER_NAME}"
# util::delete_cluster "${MEMBER_CLUSTER_1_NAME}"
# util::delete_cluster "${MEMBER_CLUSTER_2_NAME}"
