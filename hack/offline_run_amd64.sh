#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
export registry_addr_amd64=${RUNNER_NODE_IP}:${REGISTRY_PORT_AMD64}

### All AMD64 os resources###
export ARCH="amd64"
util::scope_copy_test_images ${registry_addr_amd64}
./hack/run-network-e2e.sh
./hack/offline_run_centos.sh
./hack/run-os-compatibility-e2e.sh
