#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# It should be ensured that ’kind‘ is installed in the operating environment
# Except clustername is a required option, all others are optional
CLUSTER_NAME=${1:-}
KUBECONFIG=${2:-"/root/.kube/""$CLUSTER_NAME"".config"}
KIND_VERSION=${3:-"release.daocloud.io/kpanda/kindest-node:v1.21.1"}
CONFIG_YAML=${4:-}

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${REPO_ROOT}"/hack/util.sh

# install kind and kubectl
kind_version=v0.11.1
echo -n "Preparing: 'kind' existence check - "
if util::cmd_exist kind; then
    echo "passed"
else
    echo "not pass"
    util::install_kind $kind_version
fi

if [ ! -d "/root/.kube" ]; then
    mkdir "/root/.kube"
fi


if [[ -z "${CONFIG_YAML}" ]]; then
    util::create_cluster "${CLUSTER_NAME}" "${KUBECONFIG}" "${KIND_VERSION}"
else
    util::create_cluster "${CLUSTER_NAME}" "${KUBECONFIG}" "${KIND_VERSION}" "${CONFIG_YAML}"
fi

util::check_clusters_ready "${KUBECONFIG}" "${CLUSTER_NAME}"
echo "The cluster was created successfully"
