#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
echo "ARCH: ${ARCH}"
echo "OS_NAME: ${OS_NAME}"
echo "IS_OFFLINE: ${ISOFFLINE}"

function func_power_on_vms(){
  echo "vm_name1: ${vm_name1}"
  echo "vm_name2: ${vm_name2}"
  SNAPSHOT_NAME="os-installed"
  util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
  util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"
  sleep 20
  util::wait_ip_reachable "${vm_ip_addr1}" 30
  util::wait_ip_reachable "${vm_ip_addr2}" 30
  ping -c 5 ${vm_ip_addr1}
  ping -c 5 ${vm_ip_addr2}
}

function func_prepare_config_yaml() {
    local source_path=$1
    local dest_path=$2
    rm -fr "${dest_path}"
    mkdir "${dest_path}"
    cp -f "${source_path}"/hosts-conf-cm-2nodes.yml  "${dest_path}"/hosts-conf-cm.yml
    cp -f "${source_path}"/vars-conf-cm.yml  "${dest_path}"
    cp -f "${source_path}"/kubeanCluster.yml "${dest_path}"
    cp -f "${source_path}"/kubeanClusterOps.yml  "${dest_path}"
    sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" "${dest_path}"/hosts-conf-cm.yml
    sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" "${dest_path}"/hosts-conf-cm.yml
    sed -i "s/root_password/${AMD_ROOT_PASSWORD}/g" ${dest_path}/hosts-conf-cm.yml
    sed -i "s#image:#image: ${SPRAY_JOB}#" "${dest_path}"/kubeanClusterOps.yml
}

util::vm_name_ip_init_online_by_os ${OS_NAME}
func_power_on_vms
echo "==> scp sonobuoy bin to master: "
sshpass -p "${AMD_ROOT_PASSWORD}" scp  -o StrictHostKeyChecking=no "${REPO_ROOT}"/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/

###prepare kubean install job yml using docker：kube_version: v1.24.7
source_config_path="${REPO_ROOT}"/test/common
dest_config_path="${REPO_ROOT}"/test/kubean_sonobouy_nightlye2e/e2e-install-cluster-sonobouy/
func_prepare_config_yaml "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-install-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
#* sed -i "s/containerd/docker/" "${dest_config_path}"/vars-conf-cm.yml

### prepare cluster upgrade job yml --> upgrade from v1.24.7 to v1.25.3
dest_config_path="${REPO_ROOT}"/test/kubean_sonobouy_nightlye2e/e2e-upgrade-cluster-y/
func_prepare_config_yaml "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME2="cluster1-upgrade-y"
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME2}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/cluster.yml/upgrade-cluster.yml/" "${dest_config_path}"/kubeanClusterOps.yml
#* sed -i "s/containerd/docker/" "${dest_config_path}"/vars-conf-cm.yml
sed -i "s/v1.24.7/v1.25.3/"  "${dest_config_path}"/vars-conf-cm.yml

### prepare cluster upgrade job yml --> upgrade from v1.25.3 to v1.25.5
dest_config_path="${REPO_ROOT}"/test/kubean_sonobouy_nightlye2e/e2e-upgrade-cluster-z/
func_prepare_config_yaml "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME3="cluster1-upgrade-z"
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME3}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/cluster.yml/upgrade-cluster.yml/" "${dest_config_path}"/kubeanClusterOps.yml
#* sed -i "s/containerd/docker/" "${dest_config_path}"/vars-conf-cm.yml
sed -i "s/v1.24.7/v1.25.5/"  "${dest_config_path}"/vars-conf-cm.yml

ginkgo -v -race -timeout=6h --fail-fast ./test/kubean_sonobouy_nightlye2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"

SNAPSHOT_NAME="power-down"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"

