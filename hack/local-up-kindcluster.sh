#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail
# This script starts a local kubean control plane based on current codebase and with a certain number of clusters joined.
# Parameters:
#[KUBEAN_VERSION] kubean helm-chart evision( image and helm versions should be the same)
#[IMAGE_VERSION] kubean images vision ( image and helm versions should be the same)
#[IMG_REGISTRY](optional) the image repository to be pulled from
#[HELM_REPO](optional) the helm chart repo to be pulled from
#[KIND_VERSION](optional) k8s cluster revision by specific kind image
#
# This script depends on utils in: ${REPO_ROOT}/hack/util.sh
# 1. used by developer to setup develop environment quickly.
# 2. used by e2e testing to setup test environment automatically.

KUBEAN_VERSION=${1:-latest}
KUBEAN_IMAGE_VERSION=${2:-latest}
HELM_REPO=${3:-"https://kubean-io.github.io/kubean-helm-chart"}
IMG_REGISTRY=${4:-"ghcr.io"}
KIND_VERSION=${5:-"kindest/node:v1.26.4"}
HOST_CLUSTER_NAME=${6:-"kubean-host"}

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/images.sh

# variable define
KUBECONFIG_PATH=${KUBECONFIG_PATH:-"${HOME}/.kube"}
MAIN_KUBECONFIG=${MAIN_KUBECONFIG:-"${KUBECONFIG_PATH}/${HOST_CLUSTER_NAME}.config"}
KUBEAN_SYSTEM_NAMESPACE="kubean-system"
KUBEAN_POD_LABEL="kubean"

#step0. prepare
# proxy setting in China mainland
export GOPROXY=https://goproxy.cn,direct # set domestic go proxy

# Make sure docker, go and helm exist, and the go version is a viable version.
util::cmd_must_exist "docker"
util::cmd_must_exist "go"
util::verify_go_version
util::cmd_must_exist "helm"

# install kind and kubectl
kind_version=v0.17.0
util::install_kind $kind_version

# get arch name and os name in bootstrap
BS_ARCH=$(go env GOARCH)
BS_OS=$(go env GOOS)
# check arch and os name before installing
util::install_environment_check "${BS_ARCH}" "${BS_OS}"
echo -n "Preparing: 'kubectl' existence check - "
if util::cmd_exist kubectl; then
    echo "passed"
else
    echo "not pass"
    util::install_kubectl "" "${BS_ARCH}" "${BS_OS}"
fi

# check all images were existed
IMAGE_LIST=$(images::manifest "$KUBEAN_IMAGE_VERSION" "${IMG_REGISTRY}/kubean-io")
echo "Preparing: pulling all images..."
while read -r img2pull && [[ -n "$img2pull" ]] ; do
    docker pull $img2pull
done <<< "$IMAGE_LIST"

#step1. prepare for kindClusterConfig
if [ "${OFFLINE_FLAG}" == "true" ];then
  KIND_CLUSTER_CONF_PATH="${REPO_ROOT}"/artifacts/kindClusterConfig/kubean-host-offline.yml
else
  KIND_CLUSTER_CONF_PATH="${REPO_ROOT}"/artifacts/kindClusterConfig/kubean-host.yml
fi
echo -e "Preparing kindClusterConfig in path: ${KIND_CLUSTER_CONF_PATH}"
docker pull "${KIND_VERSION}"
util::create_cluster "${HOST_CLUSTER_NAME}" "${MAIN_KUBECONFIG}" "${KIND_VERSION}" ${KIND_CLUSTER_CONF_PATH}

#step2. wait until the host cluster ready
echo "Waiting for the host clusters to be ready..."
util::check_clusters_ready "${MAIN_KUBECONFIG}" "${HOST_CLUSTER_NAME}"

#step3. load components images to kind cluster
### FIXME : below lines should be removed when kind cluster configured to resolve DNS.
#= echo "Loading images to kind cluster..."
#= while read -r img2pull && [[ -n "$img2pull" ]] ; do
    #= kind load docker-image $img2pull --name="${HOST_CLUSTER_NAME}"
#= done <<< "$IMAGE_LIST"

#step4. install kubean control plane components
echo "Installing kubean control plane components..."
export KUBECONFIG="${MAIN_KUBECONFIG}" # kube.conf for helm and kubectl

# deploy.sh (1)HELM_VER (2)IMG_VER (3)KUBE_CONF (4)TARGET_NS (5)HELM_REPO (6)IMG_REGISTRY
bash "${REPO_ROOT}"/hack/deploy.sh "${KUBEAN_VERSION}" "${KUBEAN_IMAGE_VERSION}"  "${MAIN_KUBECONFIG}"  "${KUBEAN_SYSTEM_NAMESPACE}"  "${HELM_REPO}" "${IMG_REGISTRY}" "E2E" "${REPO_ROOT}/kubean-helm-dir"

# Wait and check kubean ready
util::wait_pod_ready "${KUBEAN_POD_LABEL}" "${KUBEAN_SYSTEM_NAMESPACE}" 600s

#https://textkool.com/en/ascii-art-generator?hl=default&vl=default&font=DOS%20Rebel&text=KUBEAN
KUBEAN_GREETING='
-------------------------------------------------------------------------------------
 █████   ████ █████  █████ ███████████  ██████████   █████████   ██████   █████
░░███   ███░ ░░███  ░░███ ░░███░░░░░███░░███░░░░░█  ███░░░░░███ ░░██████ ░░███
 ░███  ███    ░███   ░███  ░███    ░███ ░███  █ ░  ░███    ░███  ░███░███ ░███
 ░███████     ░███   ░███  ░██████████  ░██████    ░███████████  ░███░░███░███
 ░███░░███    ░███   ░███  ░███░░░░░███ ░███░░█    ░███░░░░░███  ░███ ░░██████
 ░███ ░░███   ░███   ░███  ░███    ░███ ░███ ░   █ ░███    ░███  ░███  ░░█████
 █████ ░░████ ░░████████   ███████████  ██████████ █████   █████ █████  ░░█████
░░░░░   ░░░░   ░░░░░░░░   ░░░░░░░░░░░  ░░░░░░░░░░ ░░░░░   ░░░░░ ░░░░░    ░░░░░

-------------------------------------------------------------------------------------
'

function print_success() {
  echo -e "$KUBEAN_GREETING"
  echo "Local kubean is running."
  echo -e "\nTo start using your kubean, run:"
  echo -e "  export KUBECONFIG=${MAIN_KUBECONFIG}"
  echo "Please use 'kubectl config use-context kubean-host' to switch the host and control plane cluster."
}

print_success
