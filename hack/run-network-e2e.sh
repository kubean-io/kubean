#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
echo "ARCH: ${ARCH}"
echo "IS_OFFLINE: ${OFFLINE_FLAG}"

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

####################### create ipvs cluster ################
echo "create ipvs cluster....."
export OS_NAME="CENTOS7"
echo "OS_NAME: ${OS_NAME}"

util::power_on_2vms ${OS_NAME}
go_test_path="test/kubean_ipvs_cluster_e2e"
dest_config_path="${REPO_ROOT}"/${go_test_path}/${E2eInstallClusterYamlFolder}

sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
func_prepare_config_yaml_single_stack "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"
util::init_yum_repo_config_when_offline "${dest_config_path}"

CLUSTER_OPERATION_NAME1="cluster1-ipvs-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/kube_proxy_mode: iptables/kube_proxy_mode: ipvs /" "${dest_config_path}"/vars-conf-cm.yml

ginkgo -v -race -timeout=3h  --fail-fast ./${go_test_path}  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"

############## create cilium cluster ###################
if [[ "${OFFLINE_FLAG}" == "true" ]]; then
  export OS_NAME="REDHAT8"
else
  export OS_NAME="CENTOS7-HK"
fi
echo "create cilium cluster....."
echo "OS_NAME: ${OS_NAME}"

echo "Uninstall kubean ..."
helm uninstall ${LOCAL_RELEASE_NAME} -n kubean-system --kubeconfig=${KUBECONFIG_FILE}

echo "Reinstall kubean on not kubean-system ns... "
new_kubean_namespace="new-kubean-system"
bash "${REPO_ROOT}"/hack/deploy.sh "${TARGET_VERSION}" "${IMAGE_VERSION}"  "${KUBECONFIG_FILE}"  "${new_kubean_namespace}"  "${HELM_REPO}" "${IMG_REGISTRY}"
helm list -n "${new_kubean_namespace}" --kubeconfig ${KUBECONFIG_FILE}

util::power_on_2vms ${OS_NAME}
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
dest_config_path="${REPO_ROOT}"/test/kubean_cilium_cluster_e2e/e2e-install-cilium-cluster
func_prepare_config_yaml_single_stack "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"
util::init_yum_repo_config_when_offline "${dest_config_path}"

CLUSTER_OPERATION_NAME1="cluster1-cilium-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "s/kube_network_plugin: calico/kube_network_plugin: cilium/" "${dest_config_path}"/vars-conf-cm.yml
##set  kube_service_addresses: 10.88.0.0/16    kube_pods_subnet: 192.88.128.0/20
sed -i "s/10.96.0.0\/12/10.88.0.0\/16/" "${dest_config_path}"/vars-conf-cm.yml
sed -i "s/192.168.128.0/192.88.128.0/" "${dest_config_path}"/vars-conf-cm.yml
##add this line to set cilium_kube_proxy_replacement: partial, if kubespray update the cilium_kube_proxy_replacement default value to partial, this line can be deleted
sed -i "$ a\    cilium_kube_proxy_replacement: partial" "${dest_config_path}"/vars-conf-cm.yml

##set kubean operator replicas to 3
kubectl scale deployment kubean -n  ${new_kubean_namespace} --replicas=3 --kubeconfig="${KUBECONFIG_FILE}"

ginkgo -v -race -timeout=3h  --fail-fast ./test/kubean_cilium_cluster_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"

helm uninstall ${LOCAL_RELEASE_NAME} -n ${new_kubean_namespace} --kubeconfig=${KUBECONFIG_FILE}
bash "${REPO_ROOT}"/hack/deploy.sh "${TARGET_VERSION}" "${IMAGE_VERSION}"  "${KUBECONFIG_FILE}"  "kubean-system"  "${HELM_REPO}" "${IMG_REGISTRY}"
helm list -n  "kubean-system" --kubeconfig ${KUBECONFIG_FILE}
#= kubectl get pod -n kubean-system --kubeconfig=${KUBECONFIG_FILE}
############### calico dual stuck ##############
#### calico dual stack cluster need install on a Redhat8 os
#### the vm  need add a ipv6 in snapshot
if [[ "${OFFLINE_FLAG}" == "true" ]]; then
  export OS_NAME="REDHAT8"
else
  export OS_NAME="CENTOS7-HK"
fi

dest_config_path="${REPO_ROOT}"/test/kubean_calico_dualstack_e2e/e2e-install-calico-dual-stack-cluster

#### calico dual stuck: VXLAN_ALWAYS-VXLAN_ALWAYS ####

util::power_on_2vms ${OS_NAME}
func_prepare_config_yaml_dual_stack "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"
util::init_yum_repo_config_when_offline "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-vxlan-always-vxlan-always-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ipip_mode: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Always" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode_ipv6: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode_ipv6: Always" "${dest_config_path}"/vars-conf-cm.yml

sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
ginkgo -v -race --fail-fast ./test/kubean_calico_dualstack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_ALWAYS-VXLAN_ALWAYS"

#### calico dual stuck: VXLAN_CrossSubnet-VXLAN_ALWAYS ####
util::power_on_2vms ${OS_NAME}
func_prepare_config_yaml_dual_stack "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"
util::init_yum_repo_config_when_offline "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-vxlan-cross-vxlan-always-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ipip_mode: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: CrossSubnet" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode_ipv6: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode_ipv6: Always" "${dest_config_path}"/vars-conf-cm.yml
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
ginkgo -v -race --fail-fast ./test/kubean_calico_dualstack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_CrossSubnet-VXLAN_ALWAYS"

#### calico dual stuck: IPIP_ALWAYS-VXLAN_CrossSubnet ####
util::power_on_2vms ${OS_NAME}
func_prepare_config_yaml_dual_stack "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"
util::init_yum_repo_config_when_offline "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-ipip-always-vxlan-cross-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ipip_mode: Always" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_network_backend: bird" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode_ipv6: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode_ipv6: CrossSubnet" "${dest_config_path}"/vars-conf-cm.yml

sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
ginkgo -v -race --fail-fast ./test/kubean_calico_dualstack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_Always-VXLAN_CrossSubnet"

#### calico dual stuck: IPIP_CrossSubnet-VXLAN_CrossSubnet ####
util::power_on_2vms ${OS_NAME}
func_prepare_config_yaml_dual_stack "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"
util::init_yum_repo_config_when_offline "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-ipip-cross-vxlan-cross-"`date "+%H-%M-%S"`
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ipip_mode: CrossSubnet" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_network_backend: bird" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode_ipv6: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode_ipv6: CrossSubnet" "${dest_config_path}"/vars-conf-cm.yml

sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
ginkgo -v -race --fail-fast ./test/kubean_calico_dualstack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_CrossSubnet-VXLAN_CrossSubnet"

SNAPSHOT_NAME=${POWER_DOWN_SNAPSHOT_NAME}
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"

############## calico single stuck ##############
export OS_NAME="CENTOS7"
### CALICO: IPIP_ALWAYS ###
util::power_on_2vms ${OS_NAME}
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
dest_config_path="${REPO_ROOT}"/test/kubean_calico_single_stack_e2e/e2e-install-calico-cluster
func_prepare_config_yaml_single_stack "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"
util::init_yum_repo_config_when_offline "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-calico-ipip-always-"`date "+%H-%M-%S"`

sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ip_auto_method: first-found" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ip6_auto_method: first-found" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode: Always" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_network_backend: bird" "${dest_config_path}"/vars-conf-cm.yml

ginkgo -v -race --fail-fast ./test/kubean_calico_single_stack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_Always"

### CALICO: IPIP_CrossSubnet ###
util::power_on_2vms ${OS_NAME}
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
func_prepare_config_yaml_single_stack "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"
util::init_yum_repo_config_when_offline "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-calico-ipip-cross-"`date "+%H-%M-%S"`

sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ip_auto_method: kubernetes-internal-ip" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ip6_auto_method: kubernetes-internal-ip" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode:  CrossSubnet" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_network_backend: bird" "${dest_config_path}"/vars-conf-cm.yml

ginkgo -v -race --fail-fast ./test/kubean_calico_single_stack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_CrossSubnet"

### CALICO: VXLAN_ALWAYS ###
util::power_on_2vms ${OS_NAME}
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
func_prepare_config_yaml_single_stack "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"
util::init_yum_repo_config_when_offline "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-calico-vxlan-always-"`date "+%H-%M-%S"`

sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ip_auto_method: kubernetes-internal-ip" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ip6_auto_method: kubernetes-internal-ip" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode:  Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Always" "${dest_config_path}"/vars-conf-cm.yml

ginkgo -v -race --fail-fast ./test/kubean_calico_single_stack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_Always"

### CALICO: VXLAN_CrossSubnet ###
util::power_on_2vms ${OS_NAME}
sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
func_prepare_config_yaml_single_stack "${SOURCE_CONFIG_PATH}"  "${dest_config_path}"
util::init_yum_repo_config_when_offline "${dest_config_path}"
CLUSTER_OPERATION_NAME1="cluster1-calico-vxlan-cross-"`date "+%H-%M-%S"`

sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${dest_config_path}"/kubeanClusterOps.yml
sed -i "$ a\    calico_ip_auto_method: kubernetes-internal-ip" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ip6_auto_method: kubernetes-internal-ip" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode:  Never" "${dest_config_path}"/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: CrossSubnet" "${dest_config_path}"/vars-conf-cm.yml

ginkgo -v -race --fail-fast ./test/kubean_calico_single_stack_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${OFFLINE_FLAG}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_CrossSubnet"

SNAPSHOT_NAME=${POWER_DOWN_SNAPSHOT_NAME}
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"