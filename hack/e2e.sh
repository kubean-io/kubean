#!/usr/bin/env bash
set -o nounset
set -o pipefail
set -e

# This script schedules e2e tests
# Parameters:
#[TARGET_VERSION] apps ta ge images/helm-chart revision( image and helm versions should be the same)
#[IMG_REPO](optional) the image repository to be pulled from
#[HELM_REPO](optional) the helm chart repo to be pulled from
#

TARGET_VERSION=${1:-v0.0.0}
IMAGE_VERSION=${2:-latest}
HELM_REPO=${3:-"https://kubean-io.github.io/kubean-helm-chart"}
IMG_REPO=${4:-"ghcr.io/kubean-io"}
SPRAY_JOB_VERSION=${5:-latest}
RUNNER_NAME=${6:-"kubean-actions-runner1"}
EXIT_CODE=0

CLUSTER_PREFIX=kubean-"${IMAGE_VERSION}"-$RANDOM
REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${REPO_ROOT}"/hack/util.sh

local_helm_repo_alias="kubean_release"
# add kubean repo locally
repoCount=$(helm repo list | grep "${local_helm_repo_alias}" && repoCount=true || repoCount=false)
if [ "$repoCount" == "true" ]; then
    helm repo remove ${local_helm_repo_alias}
else
    echo "repoCount:" $repoCount
fi
helm repo add ${local_helm_repo_alias} ${HELM_REPO}
helm repo update
helm repo list

chmod +x ./hack/delete-cluster.sh
chmod +x ./hack/local-up-kindcluster.sh
chmod +x ./hack/run-e2e.sh
chmod +x ./hack/run-sonobouy-e2e.sh

utils:runner_ip
###### e2e logic ########
trap utils::clean_up EXIT
./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REPO}" "kindest/node:v1.21.1" "${CLUSTER_PREFIX}"-host
./hack/run-e2e.sh "${CLUSTER_PREFIX}"-host $SPRAY_JOB_VERSION $vm_ip_addr1

ret=$?
if [ ${ret} -ne 0 ]; then
  EXIT_CODE=1
fi
