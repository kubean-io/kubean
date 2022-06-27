#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# This script runs e2e test against on kubean control plane.
# You should prepare your environment in advance and following environment may be you need to set or use default one.
# - CONTROL_PLANE_KUBECONFIG: absolute path of control plane KUBECONFIG file.
#
# Usage: hack/run-e2e.sh
KUBECONFIG_PATH=${KUBECONFIG_PATH:-"${HOME}/.kube"}

HOST_CLUSTER_NAME=${1:-"kubean-host"}
MAIN_KUBECONFIG=${MAIN_KUBECONFIG:-"${KUBECONFIG_PATH}/${HOST_CLUSTER_NAME}.config"}

# Install ginkgo
GOPATH=$(go env GOPATH | awk -F ':' '{print $1}')
export PATH=$PATH:$GOPATH/bin

# Run e2e
ginkgo -v -race --fail-fast ./test/kubean_deploy_e2e/  -- --kubeconfig="${MAIN_KUBECONFIG}"


