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
export KYLIN_VM_PASSWORD=${6}
export E2E_TYPE=${7:-"PR"}


export HELM_REPO="https://kubean-io.github.io/kubean-helm-chart"
export IMG_REPO="ghcr.io/kubean-io"
export VSPHERE_HOST="10.64.56.11"
export OFFLINE_FLAG=false
export KUBECONFIG_PATH="${HOME}/.kube"
export CLUSTER_PREFIX="kubean-online-${IMAGE_VERSION}-$RANDOM"
export CONTAINERS_PREFIX="kubean-online"
export KUBECONFIG_FILE="${KUBECONFIG_PATH}/${CLUSTER_PREFIX}-host.config"
export REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh

EXIT_CODE=0
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
chmod +x ./hack/run-os-compatibility-e2e.sh

DIFF_NIGHTLYE2E=`git show -- './test/*' | grep nightlye2e || true`
DIFF_COMPATIBILE=`git show | grep /test/kubean_os_compatibility_e2e || true`

####### e2e logic ########
util::clean_online_kind_cluster

KIND_VERSION="release-ci.daocloud.io/kpanda/kindest-node:v1.25.3"
./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REPO}" "${KIND_VERSION}" "${CLUSTER_PREFIX}"-host

#utils:runner_ip

if [ "${E2E_TYPE}" == "PR" ]; then
    ./hack/run-e2e.sh
    # Judge whether to change the nightlye2e case
    if [[ -n $DIFF_NIGHTLYE2E ]] ; then
        ## pr_ci debug stage, momentarily disable nightly e2e
        echo "nightly_e2e..."
        #./hack/run-sonobouy-e2e.sh "${CLUSTER_PREFIX}"-host $SPRAY_JOB_VERSION
    fi
    # Judge whether to change the compatibility case
    if [[ -n $DIFF_COMPATIBILE ]] ; then
        ## pr_ci debug stage, momentarily disable compatibility e2e
        echo "compatibility e2e..."
        #./hack/run-os-compatibility-e2e.sh "${CLUSTER_PREFIX}"-host $SPRAY_JOB_VERSION
    fi
elif [ "${E2E_TYPE}" == "NIGHTLY" ]; then
    ./hack/run-sonobouy-e2e.sh "${CLUSTER_PREFIX}"-host $SPRAY_JOB_VERSION
else
    ./hack/run-os-compatibility-e2e.sh "${CLUSTER_PREFIX}"-host $SPRAY_JOB_VERSION
fi

util::clean_online_kind_cluster