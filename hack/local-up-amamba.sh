#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

# This script starts a local kubean control plane based on current codebase and with a certain number of clusters joined.
# Parameters:
#[KUBEAN_VERSION] kubean helm-chart evision( image and helm versions should be the same)
#[IMAGE_VERSION] kubean images vision ( image and helm versions should be the same)
#[IMG_REPO](optional) the image repository to be pulled from
#[HELM_REPO](optional) the helm chart repo to be pulled from
#[KIND_VERSION](optional) k8s cluster revision by specific kind image
#[HOST_IPADDRESS](optional) if you want to export clusters' API server port to specific IP address
#
# This script depends on utils in: ${REPO_ROOT}/hack/util.sh
# 1. used by developer to setup develop environment quickly.
# 2. used by e2e testing to setup test environment automatically.

KUBEAN_VERSION=${1:-latest}
KUBEAN_IMAGE_VERSION=${2:-latest}
HELM_REPO=${3:-"https://release.daocloud.io/chartrepo/kubean"}
IMG_REPO=${4:-"release.daocloud.io/kubean"}
KIND_REPO="release.daocloud.io"
KIND_VERSION=${5:-"${KIND_REPO}/kubean/kindest-node:v1.21.1"}
HOST_IPADDRESS=${6:-}
CLUSTERPEDIA_VERSION=${7:-"v0.1.0"}
HOST_CLUSTER_NAME=${8:-"kubean-host"}
MEMBER_CLUSTER_1_NAME=${9:-"member1"}
MEMBER_CLUSTER_2_NAME=${10:-"member2"}

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/images.sh

# variable define
KUBECONFIG_PATH=${KUBECONFIG_PATH:-"${HOME}/.kube"}
MAIN_KUBECONFIG=${MAIN_KUBECONFIG:-"${KUBECONFIG_PATH}/${HOST_CLUSTER_NAME}.config"}
MEMBER_CLUSTER_1_KUBECONFIG=${MEMBER_CLUSTER_1_KUBECONFIG:-"${KUBECONFIG_PATH}/${MEMBER_CLUSTER_1_NAME}.config"}
MEMBER_CLUSTER_2_KUBECONFIG=${MEMBER_CLUSTER_2_KUBECONFIG:-"${KUBECONFIG_PATH}/${MEMBER_CLUSTER_2_NAME}.config"}
KUBEAN_SYSTEM_NAMESPACE="kubean-system"
APISERVER_POD_LABEL="kubean-apiserver"

#step0. prepare
# proxy setting in China mainland
export GOPROXY=https://goproxy.cn,direct # set domestic go proxy

# Make sure docker, go and helm exist, and the go version is a viable version.
util::cmd_must_exist "docker"
util::cmd_must_exist "go"
util::verify_go_version
util::cmd_must_exist "helm"

# install kind and kubectl
kind_version=v0.11.1
echo -n "Preparing: 'kind' existence check - "
if util::cmd_exist kind; then
    echo "passed"
else
    echo "not pass"
    util::install_kind $kind_version
fi
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
IMAGE_LIST=$(images::manifest "$KUBEAN_IMAGE_VERSION" "$IMG_REPO")
echo "Preparing: pulling all images..."
while read -r img2pull && [[ -n "$img2pull" ]] ; do
    docker pull $img2pull
done <<< "$IMAGE_LIST"

#step1. create host cluster and member clusters in parallel
# host IP address: script parameter ahead of macOS IP
if [[ -z "${HOST_IPADDRESS}" ]]; then
  util::get_macos_ipaddress # Adapt for macOS
  HOST_IPADDRESS=${MAC_NIC_IPADDRESS:-}
fi
#prepare for kindClusterConfig
TEMP_PATH=$(mktemp -d)
echo -e "Preparing kindClusterConfig in path: ${TEMP_PATH}"
docker pull "${KIND_VERSION}"
if [[ -n "${HOST_IPADDRESS}" ]]; then # If bind the port of clusters(kubean-host, and member1) to the host IP
    cp -rf "${REPO_ROOT}"/config/kind/host.yaml "${TEMP_PATH}"/host.yaml
    sed -i'' -e "s/{{host_ipaddress}}/${HOST_IPADDRESS}/g" "${TEMP_PATH}"/host.yaml
    util::create_cluster "${HOST_CLUSTER_NAME}" "${MAIN_KUBECONFIG}" "${KIND_VERSION}" "${TEMP_PATH}"/host.yaml
else
    util::create_cluster "${HOST_CLUSTER_NAME}" "${MAIN_KUBECONFIG}" "${KIND_VERSION}"
fi

#step2. wait until the host cluster ready
echo "Waiting for the host clusters to be ready..."
util::check_clusters_ready "${MAIN_KUBECONFIG}" "${HOST_CLUSTER_NAME}"

#step4. load components images to kind cluster
### FIXME : below lines should be removed when kind cluster configured to resolve DNS.
echo "Loading images to kind cluster..."
while read -r img2pull && [[ -n "$img2pull" ]] ; do
    kind load docker-image $img2pull --name="${HOST_CLUSTER_NAME}"
done <<< "$IMAGE_LIST"

#step5. install kubean control plane components
echo "Installing kubean control plane components..."
export KUBECONFIG="${MAIN_KUBECONFIG}" # kube.conf for helm and kubectl

# deploy.sh (1)HELM_VER (2)IMG_VER (3)KUBE_CONF (4)TARGET_NS (5)HELM_REPO (6)IMG_REPO
bash "${REPO_ROOT}"/hack/deploy.sh "${KUBEAN_VERSION}" "${KUBEAN_IMAGE_VERSION}"  "${MAIN_KUBECONFIG}"  "${KUBEAN_SYSTEM_NAMESPACE}"  "${HELM_REPO}" "${IMG_REPO}" "false" "E2E"

# Wait and check kubean ready
util::wait_pod_ready "${APISERVER_POD_LABEL}" "${KUBEAN_SYSTEM_NAMESPACE}" 300s

#https://textkool.com/en/ascii-art-generator?hl=default&vl=default&font=DOS%20Rebel&text=KUBEAN
KUBEAN_GREETING='
-------------------------------------------------------------------------------------
   █████████   ██████   ██████   █████████   ██████   ██████ ███████████    █████████  
  ███░░░░░███ ░░██████ ██████   ███░░░░░███ ░░██████ ██████ ░░███░░░░░███  ███░░░░░███ 
 ░███    ░███  ░███░█████░███  ░███    ░███  ░███░█████░███  ░███    ░███ ░███    ░███ 
 ░███████████  ░███░░███ ░███  ░███████████  ░███░░███ ░███  ░██████████  ░███████████ 
 ░███░░░░░███  ░███ ░░░  ░███  ░███░░░░░███  ░███ ░░░  ░███  ░███░░░░░███ ░███░░░░░███ 
 ░███    ░███  ░███      ░███  ░███    ░███  ░███      ░███  ░███    ░███ ░███    ░███ 
 █████   █████ █████     █████ █████   █████ █████     █████ ███████████  █████   █████
░░░░░   ░░░░░ ░░░░░     ░░░░░ ░░░░░   ░░░░░ ░░░░░     ░░░░░ ░░░░░░░░░░░  ░░░░░   ░░░░░ 
                                                                                       
-------------------------------------------------------------------------------------
'

function print_success() {
  echo -e "$KUBEAN_GREETING"
  echo "Local kubean is running."
  echo -e "\nTo start using your kubean, run:"
  echo -e "  export KUBECONFIG=${MAIN_KUBECONFIG}"
  echo "Please use 'kubectl config use-context kubean-host' to switch the host and control plane cluster."
  echo -e "\nTo manage your member clusters, run:"
  echo -e "  export KUBECONFIG=${MEMBER_CLUSTER_1_KUBECONFIG}"
  echo "Please use 'kubectl config use-context member1' to switch to the member cluster."
}

print_success
