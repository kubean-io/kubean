#!/usr/bin/env bash
set -o nounset
set -o pipefail

# This script schedules e2e tests
# Parameters:
#[TARGET_VERSION] apps ta ge images/helm-chart revision( image and helm versions should be the same)
#[IMG_REPO](optional) the image repository to be pulled from
#[HELM_REPO](optional) the helm chart repo to be pulled from
#
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

###### e2e logic ########

trap clean_up EXIT
./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REPO}" "${IMG_REPO}/kindest-node:v1.21.1" "${CLUSTER_PREFIX}"-host

./hack/run-e2e.sh "${CLUSTER_PREFIX}"-host
ret=$?
if [ ${ret} -ne 0 ]; then
  EXIT_CODE=1
fi

