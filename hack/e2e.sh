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
clean_up(){
    echo 'Removing kubean kind cluster...'
    kind delete cluster --name "$CLUSTER_PREFIX"
    echo 'Done!'
}

###### e2e logic ########

trap clean_up EXIT
./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REPO}" "release.daocloud.io/kpanda/kindest-node:v1.21.1" "${CLUSTER_PREFIX}"-host

###### e2e test execution ########
ginkgo run -v -race --fail-fast ./test/e2e/

ret=$?
if [ ${ret} -ne 0 ]; then
  EXIT_CODE=1
fi
