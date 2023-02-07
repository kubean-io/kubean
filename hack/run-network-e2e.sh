#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
echo "ARCH: ${ARCH}"
echo "OS_NAME: ${OS_NAME}"
echo "IS_OFFLINE: ${ISOFFLINE}"


function func_prepare_config_yaml_dual_stack() {
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
    sed -i "$ a\    enable_dual_stack_networks: true" "${dest_config_path}"/vars-conf-cm.yml
    sed -i "$ a\    kube_pods_subnet_ipv6: fd89:ee78:d8a6:8608::1:0000/112" "${dest_config_path}"/vars-conf-cm.yml
    sed -i "$ a\    kube_service_addresses_ipv6: fd89:ee78:d8a6:8608::1000/116" "${dest_config_path}"/vars-conf-cm.yml
}

function func_prepare_config_yaml_single_stack() {
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
############## calico dual stuck ##############
### calico dual stack cluster need install on a Redhat8 os
### the vm  need add a ipv6 in snapshot
export OS_NAME="REDHAT8"
util::vm_name_ip_init_online_by_os ${OS_NAME}
echo "vm_name1: ${vm_name1}"
echo "vm_name2: ${vm_name2}"
source_config_path="${REPO_ROOT}"/test/common
dest_config_path="${REPO_ROOT}"/test/kubean_calico_dualstack_e2e/e2e-install-calico-dual-stack-cluster

### calico  VXLAN_ALWAYS-VXLAN_ALWAYS
func_prepare_config_yaml_dual_stack "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-vxlan-always-vxlan-always-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ipip_mode: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Always" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode_ipv6: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode_ipv6: Always" "${dest_config_path}"/vars-conf-cm.yml
util::power_on_2vms ${OS_NAME}
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
ginkgo -v -race --fail-fast ./test/kubean_calico_dualstack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_ALWAYS-VXLAN_ALWAYS"

### calico  VXLAN_CrossSubnet-VXLAN_ALWAYS
func_prepare_config_yaml_dual_stack "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-vxlan-cross-vxlan-always-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ipip_mode: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: CrossSubnet" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode_ipv6: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode_ipv6: Always" "${dest_config_path}"/vars-conf-cm.yml
util::power_on_2vms ${OS_NAME}
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
ginkgo -v -race --fail-fast ./test/kubean_calico_dualstack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_CrossSubnet-VXLAN_ALWAYS"

### calico  IPIP_ALWAYS-VXLAN_CrossSubnet
func_prepare_config_yaml_dual_stack "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-ipip-always-vxlan-cross-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ipip_mode: Always" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_network_backend: bird" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode_ipv6: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode_ipv6: CrossSubnet" "${dest_config_path}"/vars-conf-cm.yml

util::power_on_2vms ${OS_NAME}
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
ginkgo -v -race --fail-fast ./test/kubean_calico_dualstack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_Always-VXLAN_CrossSubnet"

### calico  IPIP_CrossSubnet-VXLAN_CrossSubnet
func_prepare_config_yaml_dual_stack "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-ipip-cross-vxlan-cross-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ipip_mode: CrossSubnet" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_network_backend: bird" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode_ipv6: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode_ipv6: CrossSubnet" "${dest_config_path}"/vars-conf-cm.yml

util::power_on_2vms ${OS_NAME}
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
ginkgo -v -race --fail-fast ./test/kubean_calico_dualstack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_CrossSubnet-VXLAN_CrossSubnet"

SNAPSHOT_NAME=${POWER_DOWN_SNAPSHOT_NAME}
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"

############## calico single stuck ##############
export OS_NAME="CENTOS7"
util::vm_name_ip_init_online_by_os ${OS_NAME}
echo "vm_name1: ${vm_name1}"
echo "vm_name2: ${vm_name2}"


### do network calico e2e
## CALICO: IPIP_ALWAYS
util::power_on_2vms ${OS_NAME}
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
source_config_path="${REPO_ROOT}"/test/common
dest_config_path="${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster
func_prepare_config_yaml_single_stack "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-calico-ipip-always-"`date "+%H-%M-%S"`

sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ip_auto_method: first-found" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ip6_auto_method: first-found" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode: Always" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_network_backend: bird" "${dest_config_path}"/vars-conf-cm.yml

ginkgo -v -race --fail-fast ./test/kubean_calico_nightlye2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_Always"

## CALICO: IPIP_CrossSubnet
util::power_on_2vms ${OS_NAME}
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
func_prepare_config_yaml_single_stack "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-calico-ipip-cross-"`date "+%H-%M-%S"`

sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ip_auto_method: kubernetes-internal-ip" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ip6_auto_method: kubernetes-internal-ip" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode:  CrossSubnet" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_network_backend: bird" "${dest_config_path}"/vars-conf-cm.yml

ginkgo -v -race --fail-fast ./test/kubean_calico_nightlye2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_CrossSubnet"

## CALICO: VXLAN_ALWAYS
util::power_on_2vms ${OS_NAME}
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
func_prepare_config_yaml_single_stack "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-calico-vxlan-always-"`date "+%H-%M-%S"`

sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ip_auto_method: kubernetes-internal-ip" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ip6_auto_method: kubernetes-internal-ip" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode:  Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Always" "${dest_config_path}"/vars-conf-cm.yml

ginkgo -v -race --fail-fast ./test/kubean_calico_nightlye2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_Always"

## CALICO: VXLAN_CrossSubnet
util::power_on_2vms ${OS_NAME}
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
func_prepare_config_yaml_single_stack "${source_config_path}"  "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-calico-vxlan-cross-"`date "+%H-%M-%S"`

sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ip_auto_method: kubernetes-internal-ip" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ip6_auto_method: kubernetes-internal-ip" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode:  Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: CrossSubnet" "${dest_config_path}"/vars-conf-cm.yml

ginkgo -v -race --fail-fast ./test/kubean_calico_nightlye2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_CrossSubnet"

SNAPSHOT_NAME=${POWER_DOWN_SNAPSHOT_NAME}
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"