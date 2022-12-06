#!/usr/bin/env bash
set -o nounset
set -o pipefail
set -e

### This script is for offline e2e
HELM_CHART_VERSION=$1
IMAGE_VERSION=$1
TAG_VERSION=$1
SPRAY_JOB_VERSION=$1
VSPHERE_USER=$2
VSPHERE_PASSWD=$3
RUNNER_NAME=$4
SPRAY_JOB="ghcr.io/kubean-io/spray-job:${SPRAY_JOB_VERSION}"
IMG_REPO="ghcr.io/kubean-io"
HELM_REPO="https://kubean-io.github.io/kubean-helm-chart/"
KUBECONFIG_PATH="${HOME}/.kube"
CLUSTER_PREFIX="kubean-offline-${IMAGE_VERSION}-$RANDOM"
KUBECONFIG_FILE="${KUBECONFIG_PATH}/${CLUSTER_PREFIX}-host.config"
OFFLINE_FLAG=true
EXIT_CODE=0
echo "HELM_CHART_VERSION: ${HELM_CHART_VERSION}"

NETWORK_CARD="ens192"
Registry_Port=31500
minIOUser="admin"
minIOPwd="adminPassword"
minioPort=32000
# Revert snapshot of vms
## Fix me: set login info to other place, such pre export to runner
VSPHERE_HOST="192.168.1.136"
SNAPSHOT_NAME="os-installed"

# Add repo if not exist; update repo if exist
REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
local_helm_repo_alias="kubean-io"
source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh

# Init vm ip and vm name value
util::vm_name_ip_init

# Offline kindCluster will be remove before test always
# Online kindCluster will be remove before test if the cpu/mem resource insufficient
# Vm snapshot will be restored before test always
CONTAINERS_PREFIX="kubean-offline"
util::clean_offline_kind_cluster
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"

# Add kubean repo locally
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
kind_ip=$(util::get_docker_native_ipaddress "${CLUSTER_PREFIX}-host-control-plane")
kubean_node_ip=$(ip a |grep ${NETWORK_CARD}|grep inet|grep global|awk -F ' ' '{print $2}'|awk -F '/' '{print $1}')
echo "Node ip of ${NETWORK_CARD} is: ${kubean_node_ip}"

### Helm install Registry: use 31500 as registry service port
util::install_registry "${Registry_Port}" "${KUBECONFIG_FILE}"

### Helm install Minio : use 32000 as minio service port
util::install_minio ${minIOUser} ${minIOPwd} "${KUBECONFIG_FILE}"
minio_ip=${kubean_node_ip}
minio_url=http://${minio_ip}:${minioPort}
echo "Minio service url: ${minio_url}"


### Download amd64 arch and uncompress tgz
arch="amd64"
DOWNLOAD_FOLDER="${REPO_ROOT}/download_offline_files-${TAG_VERSION}"
echo "Download offline files to: ${DOWNLOAD_FOLDER}"
util::download_offline_files ${arch} "${TAG_VERSION}" ${DOWNLOAD_FOLDER}
util::uncompress_tgz_files ${DOWNLOAD_FOLDER}

### Import binary files to kind minio
binarys_dir=${DOWNLOAD_FOLDER}/files
util::import_files_minio ${minIOUser} ${minIOPwd} "${minio_url}" "${binarys_dir}"

### Push images to kind registry
images_dir=${DOWNLOAD_FOLDER}/images
registry_ip=${kubean_node_ip}
registry_addr=${registry_ip}:${Registry_Port}
util::push_registry "${registry_addr}" "${images_dir}"

### Import os package
# arch type: amd64/arm64
os_package_dir=${DOWNLOAD_FOLDER}/os-pkgs
util::import_os_package_minio_by_arch ${minIOUser} ${minIOPwd} "${minio_url}" "${os_package_dir}" "${arch}"

### Import iso repos file to minio
# current linux_distribution support: "centos"
# the iso file is too big to download, so prepare it on runner
linux_distribution=centos
iso_image_file="/root/iso-images/CentOS-7-x86_64-DVD-2207-02.iso"
shell_path="${REPO_ROOT}/artifacts"
util::import_iso  ${minIOUser} ${minIOPwd} "${minio_url}" "${shell_path}" ${iso_image_file}

##### First run fundamental case in pr ci ######
## Run pr ci
CLUSTER_OPERATION_NAME1="cluster1-install-"`date +%s`

cp -f  ${REPO_ROOT}/test/offline-common/*.yml ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/

# host-config-cm.yaml set
sed -i "s/ip:/ip: ${vm_ip_addr1}/" ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/hosts-conf-cm.yml
sed -i "s/ansible_host:/ansible_host: ${vm_ip_addr1}/" ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/hosts-conf-cm.yml

# kubeanClusterOps.yml sed
sed -i "s#image:#image: ${SPRAY_JOB}#" ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/kubeanClusterOps.yml
sed -i "s#e2e-cluster1-install#${CLUSTER_OPERATION_NAME1}#" ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/kubeanClusterOps.yml
sed -i "s#{offline_minio_url}#${minio_url}#g" ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/kubeanClusterOps.yml

# vars-conf-cm.yml set
sed -i "s#registry_host:#registry_host: ${registry_addr}#"    ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/vars-conf-cm.yml
sed -i "s#minio_address:#minio_address: ${minio_url}#"    ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/vars-conf-cm.yml
sed -i "s#registry_host_key#${registry_addr}#g"    ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/vars-conf-cm.yml

# Set params in test/tools/offline_params.yml
sed -i "/ip:/c\ip: ${kubean_node_ip}"  ${REPO_ROOT}/test/tools/offline_params.yml
sed -i "/registry_addr:/c\registry_addr: ${registry_addr}"  ${REPO_ROOT}/test/tools/offline_params.yml
sed -i "/minio_addr:/c\minio_addr: ${minio_url}"  ${REPO_ROOT}/test/tools/offline_params.yml
nginx_image_name=${registry_addr}/test/$(cat /root/kubean/kubean/hack/test_images.list |grep nginx)
sed -i "/nginx_image:/c\nginx_image: ${nginx_image_name} "  ${REPO_ROOT}/test/tools/offline_params.yml

# Copy images used in test case to registry
util::scope_copy_test_images ${registry_addr}

# Run cluster function e2e
####KUBECONFIG_FILE="/root/.kube/kubean-v0.4.0-rc7-3798-host.config"
echo "Start go test ****"
ginkgo -v -race --fail-fast ./test/kubean_deploy_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}"

ginkgo -v -race -timeout=3h --fail-fast --skip "\[bug\]" ./test/kubean_functions_e2e/  -- \
           --kubeconfig="${KUBECONFIG_FILE}" \
           --clusterOperationName="${CLUSTER_OPERATION_NAME1}" --vmipaddr="${vm_ip_addr1}" --isOffline="true"

#Prepare reset yaml

CLUSTER_OPERATION_NAME2="e2e-cluster1-reset"
cp -f ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/hosts-conf-cm.yml  ${REPO_ROOT}/test/kubean_reset_e2e/e2e-reset-cluster
cp -f ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/vars-conf-cm.yml  ${REPO_ROOT}/test/kubean_reset_e2e/e2e-reset-cluster
cp -f ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/kubeanCluster.yml  ${REPO_ROOT}/test/kubean_reset_e2e/e2e-reset-cluster
sed -i "s#image:#image: ${SPRAY_JOB}#"  ${REPO_ROOT}/test/kubean_reset_e2e/e2e-reset-cluster/kubeanClusterOps.yml

# prepare kubean install job yml using docker
CLUSTER_OPERATION_NAME3="cluster1-install-dcr"`date +%s`
cp -r ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster ${REPO_ROOT}/test/kubean_reset_e2e/e2e-install-cluster-docker
sed -i "s/${CLUSTER_OPERATION_NAME1}/${CLUSTER_OPERATION_NAME3}/" ${REPO_ROOT}/test/kubean_reset_e2e/e2e-install-cluster-docker/kubeanClusterOps.yml
sed -i "s#container_manager: containerd#container_manager: docker#" ${REPO_ROOT}/test/kubean_reset_e2e/e2e-install-cluster-docker/vars-conf-cm.yml
sed -i "$ a\    override_system_hostname: false" ${REPO_ROOT}/test/kubean_reset_e2e/e2e-install-cluster-docker/vars-conf-cm.yml


ginkgo -v -race --fail-fast --skip "\[bug\]" ./test/kubean_reset_e2e/  -- \
          --kubeconfig="${KUBECONFIG_FILE}"  \
          --clusterOperationName="${CLUSTER_OPERATION_NAME3}" --vmipaddr="${vm_ip_addr1}" --isOffline="true"

## Clean up kind cluster and powerDown vm when success
util::clean_offline_kind_cluster
SNAPSHOT_NAME="power-down"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"