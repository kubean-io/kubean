#!/usr/bin/env bash
set -o nounset
set -o pipefail
set -e

# This script schedules e2e tests
# Parameters:
#[TARGET_VERSION] apps ta ge images/helm-chart revision( image and helm versions should be the same)
#[IMG_REPO](optional) the image repository to be pulled from
#[HELM_REPO](optional) the helm chart repo to be pulled from

export TARGET_VERSION=${1}
export IMAGE_VERSION=${2}
export SPRAY_JOB_VERSION=${2}
export RUNNER_NAME=${3:-"kubean-actions-runner1"}
export VSPHERE_USER=${4}
export VSPHERE_PASSWD=${5}
export AMD_ROOT_PASSWORD=${6}
export KYLIN_VM_PASSWORD=${7}
export E2E_TYPE=${8:-"PR"}
export SPRAY_JOB="ghcr.io/kubean-io/spray-job:${SPRAY_JOB_VERSION}"
export HELM_REPO="https://kubean-io.github.io/kubean-helm-chart"
export IMG_REPO="ghcr.io/kubean-io"
export VSPHERE_HOST="10.64.56.11"
export OFFLINE_FLAG=false
export KUBECONFIG_PATH="${HOME}/.kube"
export CLUSTER_PREFIX="kubean-online-$RANDOM"
export CONTAINERS_PREFIX="kubean-online"
export KUBECONFIG_FILE="${KUBECONFIG_PATH}/${CLUSTER_PREFIX}-host.config"
export REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
echo "TARGET_VERSION: ${TARGET_VERSION}"
echo "IMAGE_VERSION: ${IMAGE_VERSION}"
local_helm_repo_alias="kubean_release"

# add kubean repo locally
repoCount=true
helm repo list |awk '{print $1}'| grep "${local_helm_repo_alias}" || repoCount=false
echo "repoCount: $repoCount"
if [ "$repoCount" != "false" ]; then
    helm repo remove ${local_helm_repo_alias}
fi
helm repo add ${local_helm_repo_alias} ${HELM_REPO}
helm repo update
helm repo list

chmod +x ./hack/delete-cluster.sh
chmod +x ./hack/local-up-kindcluster.sh
chmod +x ./hack/run-e2e.sh
chmod +x ./hack/run-sonobouy-e2e.sh
chmod +x ./hack/run-os-compatibility-e2e.sh
chmod +x ./hack/run-network-e2e.sh
chmod +x ./hack/run-nightly-cluster-e2e.sh
DIFF_NIGHTLYE2E=`git show -- './test/*' | grep nightlye2e || true`
DIFF_COMPATIBILE=`git show | grep /test/kubean_os_compatibility_e2e || true`

####### e2e logic ########
util::clean_online_kind_cluster

KIND_VERSION="release-ci.daocloud.io/kpanda/kindest-node:v1.25.3"
./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REPO}" "${KIND_VERSION}" "${CLUSTER_PREFIX}"-host

if [ "${E2E_TYPE}" == "PR" ]; then
    echo "RUN PR E2E......."
    ./hack/run-e2e.sh
    # Judge whether to change the nightlye2e case
    if [[ -n $DIFF_NIGHTLYE2E ]] ; then
        echo "RUN NIGHTLY E2E......."
        ./hack/run-sonobouy-e2e.sh
    fi
    # Judge whether to change the compatibility case
    if [[ -n $DIFF_COMPATIBILE ]] ; then
        ## pr_ci debug stage, momentarily disable compatibility e2e
        echo "compatibility e2e..."
        #./hack/run-os-compatibility-e2e.sh "${CLUSTER_PREFIX}"-host $SPRAY_JOB_VERSION
    fi
elif [ "${E2E_TYPE}" == "NIGHTLY" ]; then
    echo "RUN NIGHTLY E2E......."
    ./hack/run-sonobouy-e2e.sh
else
    echo "RUN COMPATIBILITY E2E......."
    ./hack/run-os-compatibility-e2e.sh
fi

util::clean_online_kind_cluster