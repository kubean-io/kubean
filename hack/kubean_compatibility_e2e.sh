#!/usr/bin/env bash

### this script is about k8s version compatibility of kubean
set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh

function func_prepare_config_yaml_kubean_compatibility() {
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
    sed -i "s/cluster.yml/ping.yml/"  "${dest_path}"/kubeanClusterOps.yml
    #delete preHook to end lines in kubeanClusterOps.yml
    start_line=`sed -n "/preHook/=" "${dest_path}"/kubeanClusterOps.yml`
    sed -i $start_line',$d' "${dest_path}"/kubeanClusterOps.yml
}

ARCH="amd64"
OS_NAME="REDHAT8"
source_yaml_path="${REPO_ROOT}/test/common"
dest_yaml_path="${REPO_ROOT}/test/kubean_k8s_compatibility_e2e/e2e-install-cluster"
util::vm_name_ip_init_online_by_os ${OS_NAME}
echo "vm_name1: ${vm_name1}"
echo "vm_name2: ${vm_name2}"
SNAPSHOT_NAME="os-installed"
util::restore_vsphere_vm_snapshot "${VSPHERE_HOST}" "${VSPHERE_PASSWD}" "${VSPHERE_USER}" "${SNAPSHOT_NAME}" "${vm_name1}"
util::restore_vsphere_vm_snapshot "${VSPHERE_HOST}" "${VSPHERE_PASSWD}" "${VSPHERE_USER}" "${SNAPSHOT_NAME}" "${vm_name2}"
sleep 20
util::wait_ip_reachable "${vm_ip_addr1}" 10
util::wait_ip_reachable "${vm_ip_addr2}" 10
ping -c 5 "${vm_ip_addr1}"
ping -c 5 "${vm_ip_addr2}"

# prepare kubean install job yml using containerd
CLUSTER_OPERATION_NAME1="cluster1-install-"`date "+%H-%M-%S"`
echo "CLUSTER_OPERATION_NAME1: $CLUSTER_OPERATION_NAME1"
func_prepare_config_yaml_kubean_compatibility "${source_yaml_path}" "${dest_yaml_path}"
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/" "${dest_yaml_path}"/kubeanClusterOps.yml

# Run cluster function e2e
ginkgo -v -timeout=10h -race --fail-fast ./test/kubean_k8s_compatibility_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
                  --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
                  --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"
SNAPSHOT_NAME="power-down"
util::restore_vsphere_vm_snapshot "${VSPHERE_HOST}" "${VSPHERE_PASSWD}" "${VSPHERE_USER}" "${SNAPSHOT_NAME}" "${vm_name1}"
util::restore_vsphere_vm_snapshot "${VSPHERE_HOST}" "${VSPHERE_PASSWD}" "${VSPHERE_USER}" "${SNAPSHOT_NAME}" "${vm_name2}"