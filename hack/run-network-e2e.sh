#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
echo "ARCH: ${ARCH}"
echo "OS_NAME: ${OS_NAME}"
echo "IS_OFFLINE: ${ISOFFLINE}"
util::vm_name_ip_init_online_by_os ${OS_NAME}
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
#* sshpass -p root ssh -o StrictHostKeyChecking=no root@${vm_ip_addr1} cat /proc/version

### do network calico e2e
CLUSTER_OPERATION_NAME1="cluster1-calico-ipip-always-"`date "+%H-%M-%S"`
rm -fr "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
mkdir "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
cp -f "${REPO_ROOT}"/test/common/hosts-conf-cm-2nodes.yml "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
cp -f "${REPO_ROOT}"/test/common/vars-conf-cm.yml  "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
cp -f "${REPO_ROOT}"/test/common/kubeanCluster.yml "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
cp -f "${REPO_ROOT}"/test/common/kubeanClusterOps.yml  "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
sed -i "s/root_password/${AMD_ROOT_PASSWORD}/g" ${REPO_ROOT}/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/kubeanClusterOps.yml
sed -i "s#image:#image: ${SPRAY_JOB}#" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/kubeanClusterOps.yml
## CALICO: IPIP_ALWAYS
sed -i "$ a\    calico_ip_auto_method: first-found" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_ip6_auto_method: first-found" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode: Always" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Never" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_network_backend: bird" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml

ginkgo -v -race --fail-fast ./test/kubean_calico_nightlye2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_Always"

## CALICO: IPIP_CrossSubnet
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"
sleep 20
util::wait_ip_reachable "${vm_ip_addr1}" 30
util::wait_ip_reachable "${vm_ip_addr2}" 30
ping -c 5 ${vm_ip_addr1}
ping -c 5 ${vm_ip_addr2}
CLUSTER_OPERATION_NAME1="cluster1-calico-ipip-cross-"`date "+%H-%M-%S"`
rm -fr "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/*
cp -f "${REPO_ROOT}"/test/common/hosts-conf-cm-2nodes.yml "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
cp -f "${REPO_ROOT}"/test/common/vars-conf-cm.yml  "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
cp -f "${REPO_ROOT}"/test/common/kubeanCluster.yml "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
cp -f "${REPO_ROOT}"/test/common/kubeanClusterOps.yml  "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
sed -i "s/root_password/${AMD_ROOT_PASSWORD}/g" ${REPO_ROOT}/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/kubeanClusterOps.yml
sed -i "s#image:#image: ${SPRAY_JOB}#" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/kubeanClusterOps.yml

sed -i "$ a\    calico_ip_auto_method: kubernetes-internal-ip" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_ip6_auto_method: kubernetes-internal-ip" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode:  CrossSubnet" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Never" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_network_backend: bird" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml

ginkgo -v -race --fail-fast ./test/kubean_calico_nightlye2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="IPIP_CrossSubnet"

######## CALICO: VXLAN_ALWAYS
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"
sleep 20
util::wait_ip_reachable "${vm_ip_addr1}" 30
util::wait_ip_reachable "${vm_ip_addr2}" 30
ping -c 5 ${vm_ip_addr1}
ping -c 5 ${vm_ip_addr2}
CLUSTER_OPERATION_NAME1="cluster1-calico-vxlan-always-"`date "+%H-%M-%S"`
rm -fr "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/*
cp -f "${REPO_ROOT}"/test/common/hosts-conf-cm-2nodes.yml "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
cp -f "${REPO_ROOT}"/test/common/vars-conf-cm.yml  "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
cp -f "${REPO_ROOT}"/test/common/kubeanCluster.yml "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
cp -f "${REPO_ROOT}"/test/common/kubeanClusterOps.yml  "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
sed -i "s/root_password/${AMD_ROOT_PASSWORD}/g" ${REPO_ROOT}/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/kubeanClusterOps.yml
sed -i "s#image:#image: ${SPRAY_JOB}#" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/kubeanClusterOps.yml

sed -i "$ a\    calico_ip_auto_method: kubernetes-internal-ip" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_ip6_auto_method: kubernetes-internal-ip" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode:  Never" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: Always" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml

ginkgo -v -race --fail-fast ./test/kubean_calico_nightlye2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_Always"

######## CALICO: VXLAN_CrossSubnet
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"
sleep 20
util::wait_ip_reachable "${vm_ip_addr1}" 30
util::wait_ip_reachable "${vm_ip_addr2}" 30
ping -c 5 ${vm_ip_addr1}
ping -c 5 ${vm_ip_addr2}
CLUSTER_OPERATION_NAME1="cluster1-calico-vxlan-cross-"`date "+%H-%M-%S"`
rm -fr "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/*
cp -f "${REPO_ROOT}"/test/common/hosts-conf-cm-2nodes.yml "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
cp -f "${REPO_ROOT}"/test/common/vars-conf-cm.yml  "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
cp -f "${REPO_ROOT}"/test/common/kubeanCluster.yml "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
cp -f "${REPO_ROOT}"/test/common/kubeanClusterOps.yml  "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/
sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
sed -i "s/root_password/${AMD_ROOT_PASSWORD}/g" ${REPO_ROOT}/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/hosts-conf-cm.yml
sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/kubeanClusterOps.yml
sed -i "s#image:#image: ${SPRAY_JOB}#" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/kubeanClusterOps.yml

sed -i "$ a\    calico_ip_auto_method: kubernetes-internal-ip" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_ip6_auto_method: kubernetes-internal-ip" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_ipip_mode:  Never" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml
sed -i "$ a\    calico_vxlan_mode: CrossSubnet" "${REPO_ROOT}"/test/kubean_calico_nightlye2e/e2e-install-calico-cluster/vars-conf-cm.yml

ginkgo -v -race --fail-fast ./test/kubean_calico_nightlye2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
          --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
          --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"  --otherLabel="VXLAN_CrossSubnet"

SNAPSHOT_NAME="power-down"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"