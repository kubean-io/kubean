#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# This script schedules nightly e2e tests against on kubean control plane.
# Parameters:
#[TARGET_VERSION] apps ta ge images/helm-chart revision( image and helm versions should be the same)
#[IMG_REPO](optional) the image repository to be pulled from
#[HELM_REPO](optional) the helm chart repo to be pulled from
# You should prepare your environment in advance and following environment may be you need to set or use default one.
# - CONTROL_PLANE_KUBECONFIG: absolute path of control plane KUBECONFIG file.
# Usage: hack/run-nightly-e2e.sh

TARGET_VERSION=${1:-latest}
IMAGE_VERSION=${2:-latest}
HELM_REPO=${3:-"https://release.daocloud.io/chartrepo/kubean"}
IMG_REPO=${4:-"release.daocloud.io/kubean"}
EXIT_CODE=0

CLUSTER_PREFIX=kubean-"${IMAGE_VERSION}"-$RANDOM

chmod +x ./hack/delete-cluster.sh
chmod +x ./hack/local-up-kindcluster.sh
chmod +x ./hack/run-e2e.sh

###### Clean Up #######
echo "======= cluster prefix: ${CLUSTER_PREFIX}"

clean_up(){
    local auto_cleanup="true"
    if [ "$auto_cleanup" == "true" ];then
        ./hack/delete-cluster.sh "${CLUSTER_PREFIX}"-host
    fi
    if [ "$EXIT_CODE" == "0" ];then
        exit $EXIT_CODE
    fi
    exit $EXIT_CODE
}

###### nightly e2e logic ########

trap clean_up EXIT

vagrant snapshot restore default e2e_vm_initial
vagrant status

./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REPO}" "${IMG_REPO}/kindest-node:v1.21.1" "${CLUSTER_PREFIX}"-host

KUBECONFIG_PATH=${KUBECONFIG_PATH:-"${HOME}/.kube"}

HOST_CLUSTER_NAME="${CLUSTER_PREFIX}"-host
MAIN_KUBECONFIG=${MAIN_KUBECONFIG:-"${KUBECONFIG_PATH}/${HOST_CLUSTER_NAME}.config"}
# Install ginkgo
GOPATH=$(go env GOPATH | awk -F ':' '{print $1}')
export PATH=$PATH:$GOPATH/bin
# Run nightly e2e
ginkgo -v -race --fail-fast ./test/e2e/  -- --kubeconfig="${MAIN_KUBECONFIG}"

ret=$?
if [ ${ret} -ne 0 ]; then
  EXIT_CODE=1
fi






