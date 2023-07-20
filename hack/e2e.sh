#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o nounset
set -o pipefail
set -e

# This script schedules e2e tests
# Parameters:
#[TARGET_VERSION] apps ta ge images/helm-chart revision( image and helm versions should be the same)
#[IMG_REGISTRY](optional) the image repository to be pulled from
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
export IMG_REGISTRY="ghcr.m.daocloud.io"
export VSPHERE_HOST="10.64.56.11"
export OFFLINE_FLAG=false
export KUBECONFIG_PATH="${HOME}/.kube"
export CLUSTER_PREFIX="kubean-online-$RANDOM"
export CONTAINERS_PREFIX="kubean-online"
export KUBECONFIG_FILE="${KUBECONFIG_PATH}/${CLUSTER_PREFIX}-host.config"
export REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
export POWER_ON_SNAPSHOT_NAME="os-installed"
export POWER_DOWN_SNAPSHOT_NAME="power-down"
export LOCAL_REPO_ALIAS="kubean_release"
export LOCAL_RELEASE_NAME=kubean
export E2eInstallClusterYamlFolder="e2e-install-cluster"

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
echo "TARGET_VERSION: ${TARGET_VERSION}"
echo "IMAGE_VERSION: ${IMAGE_VERSION}"

# add kubean repo locally
repoCount=true
helm repo list |awk '{print $1}'| grep "${LOCAL_REPO_ALIAS}" || repoCount=false
echo "repoCount: $repoCount"
if [ "$repoCount" != "false" ]; then
    helm repo remove ${LOCAL_REPO_ALIAS}
fi
helm repo add ${LOCAL_REPO_ALIAS} ${HELM_REPO}
helm repo update
helm repo list

chmod +x ./hack/delete-cluster.sh
chmod +x ./hack/local-up-kindcluster.sh
chmod +x ./hack/run-e2e.sh
chmod +x ./hack/run-sonobouy-e2e.sh
chmod +x ./hack/run-os-compatibility-e2e.sh
chmod +x ./hack/run-network-e2e.sh
chmod +x ./hack/run-nightly-cluster-e2e.sh
chmod +x ./hack/kubean_compatibility_e2e.sh
chmod +x ./hack/kubean_resource.sh
chmod +x ./hack/autoversion.sh
chmod +x ./hack/run-vip.sh
DIFF_NIGHTLYE2E=`git show -- './test/*' | grep nightlye2e || true`
DIFF_COMPATIBILE=`git show | grep /test/kubean_os_compatibility_e2e || true`

####### e2e logic ########
if [ "${E2E_TYPE}" == "KUBEAN-COMPATIBILITY" ]; then
    k8s_list=( "v1.20.15" "v1.21.14" "v1.22.15" "v1.23.13" "v1.24.7" "v1.25.3" "v1.26.0" "v1.27.1" )
    echo ${#k8s_list[@]}
    for k8s in "${k8s_list[@]}"; do
        echo "***************k8s version is: ${k8s} ***************"
        kind::clean_kind_cluster ${CONTAINERS_PREFIX}
        KIND_VERSION="release-ci.daocloud.io/kpanda/kindest-node:"${k8s}
        ./hack/autoversion.sh "${IMAGE_VERSION}"
        ./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REGISTRY}" "${KIND_VERSION}" "${CLUSTER_PREFIX}"-host
        ./hack/kubean_compatibility_e2e.sh
    done

else
    kind::clean_kind_cluster ${CONTAINERS_PREFIX}
    KIND_VERSION="release-ci.daocloud.io/kpanda/kindest-node:v1.26.4"
    ./hack/local-up-kindcluster.sh "${TARGET_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REGISTRY}" "${KIND_VERSION}" "${CLUSTER_PREFIX}"-host
    util::set_config_path
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

fi

kind::clean_kind_cluster ${CONTAINERS_PREFIX}