#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source "${REPO_ROOT}"/hack/util.sh
source "${REPO_ROOT}"/hack/offline-util.sh
export ISOFFLINE=false
echo "IS_OFFLINE: ${ISOFFLINE}"

os_compability_e2e(){
    OS_NAME=${1}
    ARCH="amd64"
    util::vm_name_ip_init_online_by_os ${OS_NAME}
    echo "vm_name1: ${vm_name1}"
    echo "vm_name2: ${vm_name2}"
    SNAPSHOT_NAME="os-installed"
    util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
    util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"
    sleep 20
    util::wait_ip_reachable "${vm_ip_addr1}" 10
    util::wait_ip_reachable "${vm_ip_addr2}" 10
    ping -c 5 ${vm_ip_addr1}
    ping -c 5 ${vm_ip_addr2}

    rm -f ~/.ssh/known_hosts
    echo "==> scp sonobuoy bin to master: "
    sshpass -p ${AMD_ROOT_PASSWORD} scp -o StrictHostKeyChecking=no ${REPO_ROOT}/test/tools/sonobuoy root@$vm_ip_addr1:/usr/bin/
    sshpass -p ${AMD_ROOT_PASSWORD} ssh root@$vm_ip_addr1 "chmod +x /usr/bin/sonobuoy"

    # prepare kubean install job yml using containerd
    CLUSTER_OPERATION_NAME1="cluster1-install-"`date "+%H-%M-%S"`
    echo "CLUSTER_OPERATION_NAME1: $CLUSTER_OPERATION_NAME1"
    rm -fr "${REPO_ROOT}"/test/kubean_os_compatibility_e2e/e2e-install-cluster/
    mkdir "${REPO_ROOT}"/test/kubean_os_compatibility_e2e/e2e-install-cluster/
    cp -f "${REPO_ROOT}"/test/common/hosts-conf-cm-2nodes.yml "${REPO_ROOT}"/test/kubean_os_compatibility_e2e/e2e-install-cluster/hosts-conf-cm.yml
    cp -f "${REPO_ROOT}"/test/common/vars-conf-cm.yml  "${REPO_ROOT}"/test/kubean_os_compatibility_e2e/e2e-install-cluster/
    cp -f "${REPO_ROOT}"/test/common/kubeanCluster.yml "${REPO_ROOT}"/test/kubean_os_compatibility_e2e/e2e-install-cluster/
    cp -f "${REPO_ROOT}"/test/common/kubeanClusterOps.yml  "${REPO_ROOT}"/test/kubean_os_compatibility_e2e/e2e-install-cluster/
    sed -i "s/vm_ip_addr1/${vm_ip_addr1}/" ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/hosts-conf-cm.yml
    sed -i "s/vm_ip_addr2/${vm_ip_addr2}/" ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/hosts-conf-cm.yml
    sed -i "s/root_password/${AMD_ROOT_PASSWORD}/g" ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/hosts-conf-cm.yml
    sed -i "s/e2e-cluster1-install/${CLUSTER_OPERATION_NAME1}/"  ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/kubeanClusterOps.yml
    sed -i "s#image:.*#image: ${SPRAY_JOB}#" ${REPO_ROOT}/test/kubean_os_compatibility_e2e/e2e-install-cluster/kubeanClusterOps.yml
    # Run cluster function e2e
    ginkgo -v -timeout=10h -race --fail-fast ./test/kubean_os_compatibility_e2e/  -- --kubeconfig="${KUBECONFIG_FILE}" \
                      --clusterOperationName="${CLUSTER_OPERATION_NAME1}"  --vmipaddr="${vm_ip_addr1}" --vmipaddr2="${vm_ip_addr2}" \
                      --isOffline="${ISOFFLINE}" --arch=${ARCH}  --vmPassword="${AMD_ROOT_PASSWORD}"

    SNAPSHOT_NAME="power-down"
    util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name1}"
    util::restore_vsphere_vm_snapshot ${VSPHERE_HOST} ${VSPHERE_PASSWD} ${VSPHERE_USER} "${SNAPSHOT_NAME}" "${vm_name2}"

}

###### OS compitable e2e logic ########
os_array=("REDHAT7" "REDHAT8")
echo "OS list: ${os_array[*]}"
for (( i=0; i<${#os_array[@]};i++)); do
    echo "***************"
    echo "OS is: ${os_array[$i]}"
    os_compability_e2e ${os_array[$i]}
done



