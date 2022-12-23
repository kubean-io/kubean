#!/usr/bin/env bash
set -o nounset
set -o pipefail
set -e

### This script is for offline e2e
HELM_CHART_VERSION=$1
IMAGE_VERSION=$1
export SPRAY_JOB_VERSION=$1
export TAG_VERSION=$1
export VSPHERE_USER=$2
export VSPHERE_PASSWD=$3
export KYLIN_VM_PASSWORD=$4
export VSPHERE_HOST="10.64.56.11"
export RUNNER_NAME=$5
export SPRAY_JOB="m.daocloud.io/ghcr.io/kubean-io/spray-job:${SPRAY_JOB_VERSION}"
export REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
IMG_REPO="ghcr.io/kubean-io"
HELM_REPO="https://kubean-io.github.io/kubean-helm-chart/"
KUBECONFIG_PATH="${HOME}/.kube"
CLUSTER_PREFIX="kubean-offline-${IMAGE_VERSION}-$RANDOM"
export KUBECONFIG_FILE="${KUBECONFIG_PATH}/${CLUSTER_PREFIX}-host.config"
export OFFLINE_FLAG=true
EXIT_CODE=0

export REGISTRY_PORT_AMD64=31500
export REGISTRY_PORT_ARM64=31501
export MINIOUSER="admin"
export MINIOPWD="adminPassword"
export MINIOPORT=32000

export CONTAINERS_PREFIX="kubean-offline"
export DOWNLOAD_FOLDER="${REPO_ROOT}/download_offline_files-${TAG_VERSION}"
export REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
echo "HELM_CHART_VERSION: ${HELM_CHART_VERSION}"
NETWORK_CARD="ens192"
export RUNNER_NODE_IP=$(ip a |grep ${NETWORK_CARD}|grep inet|grep global|awk -F ' ' '{print $2}'|awk -F '/' '{print $1}')
export MINIO_URL=http://${RUNNER_NODE_IP}:${MINIOPORT}

chmod +x ${REPO_ROOT}/hack/offline_run_amd64.sh
chmod +x ${REPO_ROOT}/hack/offline_run_arm64.sh
chmod +x ${REPO_ROOT}/hack/offline_run_centos.sh

registry_addr_amd64=${RUNNER_NODE_IP}:${REGISTRY_PORT_AMD64}
registry_addr_arm64=${RUNNER_NODE_IP}:${REGISTRY_PORT_ARM64}
local_helm_repo_alias="kubean-io"
source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh

util::clean_offline_kind_cluster
repoCount=$(helm repo list | grep "${local_helm_repo_alias}" && repoCount=true || repoCount=false)
if [ "$repoCount" == "true" ]; then
  helm repo remove ${local_helm_repo_alias}
else
   echo "repoCount:" $repoCount
fi
helm repo add ${local_helm_repo_alias} ${HELM_REPO}
helm repo update
helm repo list

KIND_VERSION="release-ci.daocloud.io/kpanda/kindest-node:v1.25.3"
./hack/local-up-kindcluster.sh "${HELM_CHART_VERSION}" "${IMAGE_VERSION}" "${HELM_REPO}" "${IMG_REPO}" "${KIND_VERSION}" "${CLUSTER_PREFIX}"-host

### Helm install Registry: AMD64 registry and ARM64 registry
util::install_registry "${REGISTRY_PORT_AMD64}" "${KUBECONFIG_FILE}" "registry-amd64"
util::install_registry "${REGISTRY_PORT_ARM64}" "${KUBECONFIG_FILE}" "registry-arm64"
### Helm install Minio
util::install_minio ${MINIOUSER} ${MINIOPWD} "${KUBECONFIG_FILE}"

### Download release files
util::download_offline_files_by_tag "${TAG_VERSION}" ${DOWNLOAD_FOLDER}

### Set params in test/tools/offline_params.yml
sed -i "/ip:/c\ip: ${RUNNER_NODE_IP}"  ${REPO_ROOT}/test/tools/offline_params.yml
sed -i "/registry_addr_amd64:/c\registry_addr_amd64: ${registry_addr_amd64}"  ${REPO_ROOT}/test/tools/offline_params.yml
sed -i "/registry_addr_arm64:/c\registry_addr_arm64: ${registry_addr_arm64}"  ${REPO_ROOT}/test/tools/offline_params.yml
sed -i "/minio_addr:/c\minio_addr: ${MINIO_URL}"  ${REPO_ROOT}/test/tools/offline_params.yml
nginx_image_name="${registry_addr_amd64}/test/docker.m.daocloud.io/library/nginx:1.22"
sed -i "/nginx_image_amd64:/c\nginx_image_amd64: ${nginx_image_name} "  ${REPO_ROOT}/test/tools/offline_params.yml
nginx_image_name="${registry_addr_arm64}/test/docker.m.daocloud.io/arm64v8/nginx:1.23-alpine"
sed -i "/nginx_image_arm64:/c\nginx_image_arm64: ${nginx_image_name} "  ${REPO_ROOT}/test/tools/offline_params.yml

##### First AMD64 case ######
./hack/offline_run_amd64.sh
./hack/offline_run_arm64.sh

util::clean_offline_kind_cluster