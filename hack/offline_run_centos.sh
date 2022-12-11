#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh

## Resource prepare
os_name="centos7"
shell_path="${REPO_ROOT}/artifacts"
iso_image_file="/root/iso-images/CentOS-7-x86_64-DVD-2207-02.iso"

util::import_os_package_minio ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${DOWNLOAD_FOLDER}" "${os_name}"
util::import_iso  ${MINIOUSER} ${MINIOPWD} "${MINIO_URL}" "${shell_path}" ${iso_image_file}

############### Base case
CLUSTER_OPERATION_NAME1="cluster1-install-"`date "+%H-%M-%S"`
cp -f  ${REPO_ROOT}/test/offline-common/hosts-conf-cm.yml ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/
cp -f  ${REPO_ROOT}/test/offline-common/kubeanCluster.yml ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/
cp -f  ${REPO_ROOT}/test/offline-common/kubeanClusterOps.yml ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/
cp -f  ${REPO_ROOT}/test/offline-common/vars-conf-cm.yml ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/

util::vm_name_ip_init_offline_by_os ${os_name}
# host-config-cm.yaml set
sed -i "s/ip:/ip: ${vm_ip_addr1}/" ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/hosts-conf-cm.yml
sed -i "s/ansible_host:/ansible_host: ${vm_ip_addr1}/" ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/hosts-conf-cm.yml

# kubeanClusterOps.yml sed
sed -i "s#image:#image: ${SPRAY_JOB}#" ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/kubeanClusterOps.yml
sed -i "s#e2e-cluster1-install#${CLUSTER_OPERATION_NAME1}#" ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/kubeanClusterOps.yml
sed -i "s#{offline_minio_url}#${MINIO_URL}#g" ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/kubeanClusterOps.yml

# vars-conf-cm.yml set
sed -i "s#registry_host:#registry_host: ${registry_addr_amd64}#"    ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/vars-conf-cm.yml
sed -i "s#minio_address:#minio_address: ${MINIO_URL}#"    ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/vars-conf-cm.yml
sed -i "s#registry_host_key#${registry_addr_amd64}#g"    ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/vars-conf-cm.yml

# restore vm snapshot
SNAPSHOT_NAME="os-installed"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
echo "GINKGO VERSISON:"
ginkgo version
ginkgo -v -race --fail-fast ./test/kubean_deploy_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}"

ginkgo -v -race -timeout=3h --fail-fast --skip "\[bug\]" ./test/kubean_functions_e2e/  -- \
          --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}" --vmipaddr="${vm_ip_addr1}" --isOffline="true" --arch=${arch}

# Prepare reset yaml
CLUSTER_OPERATION_NAME2="e2e-cluster1-reset"
cp -f ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/hosts-conf-cm.yml  ${REPO_ROOT}/test/kubean_reset_e2e/e2e-reset-cluster
cp -f ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/vars-conf-cm.yml  ${REPO_ROOT}/test/kubean_reset_e2e/e2e-reset-cluster
cp -f ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster/kubeanCluster.yml  ${REPO_ROOT}/test/kubean_reset_e2e/e2e-reset-cluster
sed -i "s#image:#image: ${SPRAY_JOB}#"  ${REPO_ROOT}/test/kubean_reset_e2e/e2e-reset-cluster/kubeanClusterOps.yml

# Prepare kubean install job yml using docker
CLUSTER_OPERATION_NAME3="cluster1-install-dcr"`date "+%H-%M-%S"`
cp -r ${REPO_ROOT}/test/kubean_functions_e2e/e2e-install-cluster ${REPO_ROOT}/test/kubean_reset_e2e/e2e-install-cluster-docker
sed -i "s/${CLUSTER_OPERATION_NAME1}/${CLUSTER_OPERATION_NAME3}/" ${REPO_ROOT}/test/kubean_reset_e2e/e2e-install-cluster-docker/kubeanClusterOps.yml
sed -i "s#container_manager: containerd#container_manager: docker#" ${REPO_ROOT}/test/kubean_reset_e2e/e2e-install-cluster-docker/vars-conf-cm.yml
sed -i "$ a\    override_system_hostname: false" ${REPO_ROOT}/test/kubean_reset_e2e/e2e-install-cluster-docker/vars-conf-cm.yml
echo "in shell arch is ${arch}"
ginkgo -v -race --fail-fast --skip "\[bug\]" ./test/kubean_reset_e2e/  -- \
          --kubeconfig="${KUBECONFIG_FILE}"  \
          --clusterOperationName="${CLUSTER_OPERATION_NAME3}" --vmipaddr="${vm_ip_addr1}" --isOffline="true" --arch=${arch}

SNAPSHOT_NAME="power-down"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"