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

TARGET_VERSION=${1:-v0.0.1}
IMAGE_VERSION=${2:-latest}
HELM_REPO=${3:-"https://kubean-io.github.io/kubean-helm-chart"}
IMG_REPO=${4:-v0.0.1}
SPRAY_JOB_VERSION=${5:-"ghcr.io/kubean-io/kubean/spray-job:v0.0.1"}
RUNNER_NAME=${6:-"kubean-actions-runner1"} 
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

###### to get k8 cluster single node ip address based on actions-runner #######
echo "RUNNER_NAME: "$RUNNER_NAME
if [ "${RUNNER_NAME}" == "kubean-actions-runner1" ]; then
    vm_ip_addr="10.6.127.33"
fi
if [ "${RUNNER_NAME}" == "kubean-actions-runner2" ]; then
    vm_ip_addr="10.6.127.35"
fi
if [ "${RUNNER_NAME}" == "debug" ]; then
    vm_ip_addr="10.6.127.41"
fi

###### e2e logic ########
trap clean_up EXIT
./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REPO}" "kindest/node:v1.21.1" "${CLUSTER_PREFIX}"-host
./hack/run-e2e.sh "${CLUSTER_PREFIX}"-host $SPRAY_JOB_VERSION $vm_ip_addr

ret=$?
if [ ${ret} -ne 0 ]; then
  EXIT_CODE=1
fi